package analysis

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// refKey deduplicates references so the same source location is not counted
// twice when it appears in both a regular package and its test-compiled variant.
type refKey struct {
	pos token.Pos
	env Env
}

// interfaceKey identifies an interface type across package variants.  Pointer
// identity is not stable when the same source package is loaded in regular and
// test-compiled forms, so named interfaces are keyed by their defining object.
type interfaceKey struct {
	pkgPath string
	name    string
}

// ifaceAssignment records that a concrete named type was observed being used
// as a particular interface type, enabling method-usage propagation.
type ifaceAssignment struct {
	concrete *types.Named
	iface    interfaceKey
}

// ifaceMethodUse records that a method was selected on an interface value.
type ifaceMethodUse struct {
	iface  interfaceKey
	method string
	env    Env
}

type interfaceDispatchFacts struct {
	assigns    []ifaceAssignment
	methodUses []ifaceMethodUse
}

// collectReferences walks all reference packages, marks used declarations in
// the UsageIndex, and returns interface-dispatch facts found in the module.
func collectReferences(
	pkgs []*packages.Package,
	fset *token.FileSet,
	declIdx *DeclIndex,
	usage *UsageIndex,
) interfaceDispatchFacts {
	seen := make(map[refKey]bool)
	var facts interfaceDispatchFacts

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		collectPkgReferences(pkg, fset, declIdx, usage, seen, &facts)
	}
	return facts
}

func collectPkgReferences(
	pkg *packages.Package,
	fset *token.FileSet,
	declIdx *DeclIndex,
	usage *UsageIndex,
	seen map[refKey]bool,
	facts *interfaceDispatchFacts,
) {
	info := pkg.TypesInfo

	// Build a set of identifier positions that appear as receiver types inside
	// method declarations.  e.g. "func (T) Method()" — T refers to the type
	// but is not a genuine consumer usage; removing T would also remove Method.
	skipPos := make(map[token.Pos]bool)
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				return true
			}
			for _, field := range fn.Recv.List {
				collectTypeIdentPositions(field.Type, skipPos)
			}
			return true
		})
	}

	// ── Direct identifier references ────────────────────────────────────────
	// types.Info.Uses maps every identifier (including selector names) to the
	// object it resolves to, so this captures functions, types, variables,
	// constants, methods, fields, and interface methods in one pass.
	for ident, obj := range info.Uses {
		if _, ok := declIdx.get(obj); !ok {
			continue
		}
		// Skip the declaration site itself.
		if obj.Pos() == ident.Pos() {
			continue
		}
		// Skip receiver-type references inside method declarations.
		if skipPos[ident.NamePos] {
			continue
		}
		env := fileEnv(fset.Position(ident.Pos()).Filename)
		key := refKey{pos: ident.NamePos, env: env}
		if seen[key] {
			continue
		}
		seen[key] = true
		usage.Mark(obj, env)
	}

	// ── Promoted selections — mark intermediate embedded fields ──────────────
	// For x.f where f is promoted through one or more anonymous fields,
	// types.Info.Uses only records the final field/method.  We also need to
	// mark every anonymous field along the promotion path as used, so that
	// "embedded via struct" does not appear as unused.
	for selExpr, sel := range info.Selections {
		if fn, ok := sel.Obj().(*types.Func); ok {
			if key, ok := interfaceKeyForType(sel.Recv()); ok {
				env := fileEnv(fset.Position(selExpr.Pos()).Filename)
				facts.methodUses = append(facts.methodUses, ifaceMethodUse{
					iface:  key,
					method: fn.Name(),
					env:    env,
				})
			}
		}

		indices := sel.Index()
		if len(indices) <= 1 {
			continue // not promoted; already handled by Uses
		}
		env := fileEnv(fset.Position(selExpr.Pos()).Filename)
		recvType := sel.Recv()
		for _, idx := range indices[:len(indices)-1] {
			if ptr, ok := recvType.(*types.Pointer); ok {
				recvType = ptr.Elem()
			}
			st, ok := recvType.Underlying().(*types.Struct)
			if !ok {
				break
			}
			if idx >= st.NumFields() {
				break
			}
			field := st.Field(idx)
			if _, ok := declIdx.get(field); ok {
				key := refKey{pos: selExpr.Pos(), env: env}
				if !seen[key] {
					seen[key] = true
					usage.Mark(field, env)
				}
			}
			recvType = field.Type()
		}
	}

	// ── Interface assignments — collect for dispatch propagation ─────────────
	for _, file := range pkg.Syntax {
		collectIfaceAssignsInFile(file, info, &facts.assigns)
	}
}

func collectIfaceAssignsInFile(file *ast.File, info *types.Info, assigns *[]ifaceAssignment) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body == nil {
				return false
			}
			obj := info.Defs[node.Name]
			if obj == nil {
				return false
			}
			sig, ok := obj.Type().Underlying().(*types.Signature)
			if !ok {
				return false
			}
			collectIfaceAssignsInFunc(node.Body, sig, info, assigns)
			return false
		case *ast.FuncLit:
			sig, ok := info.TypeOf(node).Underlying().(*types.Signature)
			if !ok {
				return false
			}
			collectIfaceAssignsInFunc(node.Body, sig, info, assigns)
			return false
		default:
			collectIfaceAssigns(n, nil, info, assigns)
			return true
		}
	})
}

func collectIfaceAssignsInFunc(
	body *ast.BlockStmt,
	sig *types.Signature,
	info *types.Info,
	assigns *[]ifaceAssignment,
) {
	ast.Inspect(body, func(n ast.Node) bool {
		if lit, ok := n.(*ast.FuncLit); ok {
			litSig, ok := info.TypeOf(lit).Underlying().(*types.Signature)
			if !ok {
				return false
			}
			collectIfaceAssignsInFunc(lit.Body, litSig, info, assigns)
			return false
		}
		collectIfaceAssigns(n, sig.Results(), info, assigns)
		return true
	})
}

// collectIfaceAssigns inspects one AST node and records any concrete→interface
// assignments it finds.
func collectIfaceAssigns(n ast.Node, results *types.Tuple, info *types.Info, assigns *[]ifaceAssignment) {
	switch node := n.(type) {

	case *ast.AssignStmt:
		for i, rhs := range node.Rhs {
			if i >= len(node.Lhs) {
				break
			}
			lhsT := info.TypeOf(node.Lhs[i])
			rhsT := info.TypeOf(rhs)
			recordIfaceAssign(lhsT, rhsT, assigns)
		}

	case *ast.ValueSpec:
		if node.Type != nil {
			specT := info.TypeOf(node.Type)
			for _, val := range node.Values {
				recordIfaceAssign(specT, info.TypeOf(val), assigns)
			}
		}

	case *ast.CallExpr:
		fnT := info.TypeOf(node.Fun)
		if fnT == nil {
			return
		}
		sig, ok := fnT.Underlying().(*types.Signature)
		if !ok {
			return
		}
		params := sig.Params()
		if params.Len() == 0 {
			return
		}
		for i, arg := range node.Args {
			var paramT types.Type
			if sig.Variadic() && i >= params.Len()-1 {
				if sliceT, ok := params.At(params.Len() - 1).Type().(*types.Slice); ok {
					paramT = sliceT.Elem()
				}
			} else if i < params.Len() {
				paramT = params.At(i).Type()
			}
			if paramT == nil {
				continue
			}
			recordIfaceAssign(paramT, info.TypeOf(arg), assigns)
		}

	case *ast.ReturnStmt:
		if results == nil || results.Len() != len(node.Results) {
			return
		}
		for i, result := range node.Results {
			recordIfaceAssign(results.At(i).Type(), info.TypeOf(result), assigns)
		}

	case *ast.CompositeLit:
		litT := info.TypeOf(node)
		if litT == nil {
			return
		}
		st, ok := litT.Underlying().(*types.Struct)
		if !ok {
			return
		}
		for i, elt := range node.Elts {
			switch e := elt.(type) {
			case *ast.KeyValueExpr:
				if key, ok := e.Key.(*ast.Ident); ok {
					for j := range st.NumFields() {
						if st.Field(j).Name() == key.Name {
							recordIfaceAssign(st.Field(j).Type(), info.TypeOf(e.Value), assigns)
							break
						}
					}
				}
			default:
				if i < st.NumFields() {
					recordIfaceAssign(st.Field(i).Type(), info.TypeOf(elt), assigns)
				}
			}
		}
	}
}

// recordIfaceAssign records a concrete→interface pair when lhsT is an
// interface and rhsT is (or dereferences to) a named concrete type.
func recordIfaceAssign(lhsT, rhsT types.Type, assigns *[]ifaceAssignment) {
	if lhsT == nil || rhsT == nil {
		return
	}
	iface, ok := interfaceKeyForType(lhsT)
	if !ok {
		return
	}
	named := concreteNamed(rhsT)
	if named == nil {
		return
	}
	*assigns = append(*assigns, ifaceAssignment{concrete: named, iface: iface})
}

// concreteNamed extracts the *types.Named from a type, dereferencing at most
// one pointer level.  Returns nil for interface types.
func concreteNamed(t types.Type) *types.Named {
	t = types.Unalias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	if _, isIface := named.Underlying().(*types.Interface); isIface {
		return nil
	}
	return named
}

func interfaceKeyForType(t types.Type) (interfaceKey, bool) {
	t = types.Unalias(t)
	named, ok := t.(*types.Named)
	if !ok {
		return interfaceKey{}, false
	}
	if _, ok := named.Underlying().(*types.Interface); !ok {
		return interfaceKey{}, false
	}
	obj := named.Obj()
	if obj == nil {
		return interfaceKey{}, false
	}
	pkgPath := ""
	if obj.Pkg() != nil {
		pkgPath = obj.Pkg().Path()
	}
	return interfaceKey{pkgPath: pkgPath, name: obj.Name()}, true
}

// propagateInterfaceDispatch propagates usage from used interface methods to
// the corresponding concrete methods on types that were assigned to those
// interfaces.
func propagateInterfaceDispatch(
	declIdx *DeclIndex,
	usage *UsageIndex,
	facts interfaceDispatchFacts,
) {
	if len(facts.assigns) == 0 || len(facts.methodUses) == 0 {
		return
	}

	ifaceToConcretes := make(map[interfaceKey][]*types.Named)
	for _, a := range facts.assigns {
		ifaceToConcretes[a.iface] = append(ifaceToConcretes[a.iface], a.concrete)
	}

	for _, use := range facts.methodUses {
		for _, concrete := range ifaceToConcretes[use.iface] {
			markConcreteMethod(concrete, use.method, use.env, declIdx, usage)
		}
	}
}

// markConcreteMethod looks for methodName on named (value or pointer receiver)
// and marks it as used when it is a tracked declaration.
func markConcreteMethod(
	named *types.Named,
	methodName string,
	env Env,
	declIdx *DeclIndex,
	usage *UsageIndex,
) {
	for _, candidate := range []types.Type{
		named,
		types.NewPointer(named),
	} {
		mset := types.NewMethodSet(candidate)
		sel := mset.Lookup(nil, methodName)
		if sel == nil {
			continue
		}
		fn, ok := sel.Obj().(*types.Func)
		if !ok {
			continue
		}
		if _, declared := declIdx.get(fn); declared {
			usage.Mark(fn, env)
		}
	}
}

// collectTypeIdentPositions traverses a type expression and records the
// token.Pos of every identifier it contains into pos.
// Used to identify receiver-type references that must not be counted as usage.
func collectTypeIdentPositions(expr ast.Expr, pos map[token.Pos]bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Ident:
		pos[e.NamePos] = true
	case *ast.StarExpr:
		collectTypeIdentPositions(e.X, pos)
	case *ast.IndexExpr:
		collectTypeIdentPositions(e.X, pos)
		collectTypeIdentPositions(e.Index, pos)
	case *ast.IndexListExpr:
		collectTypeIdentPositions(e.X, pos)
		for _, idx := range e.Indices {
			collectTypeIdentPositions(idx, pos)
		}
	case *ast.SelectorExpr:
		collectTypeIdentPositions(e.X, pos)
		pos[e.Sel.NamePos] = true
	case *ast.ParenExpr:
		collectTypeIdentPositions(e.X, pos)
	}
}
