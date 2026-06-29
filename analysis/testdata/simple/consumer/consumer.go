package consumer

import (
	"example.com/simple/internal/pkg"
)

func Use() {
	pkg.UsedFunc()
	_ = pkg.UsedType{UsedField: "x"}
	_ = pkg.UsedVar
	_ = pkg.UsedConst
}
