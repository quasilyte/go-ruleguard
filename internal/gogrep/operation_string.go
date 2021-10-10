// Code generated by "stringer -type=operation -trimprefix=op"; DO NOT EDIT.

package gogrep

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[opInvalid-0]
	_ = x[opNode-1]
	_ = x[opNamedNode-2]
	_ = x[opNodeSeq-3]
	_ = x[opNamedNodeSeq-4]
	_ = x[opOptNode-5]
	_ = x[opNamedOptNode-6]
	_ = x[opMultiStmt-7]
	_ = x[opMultiExpr-8]
	_ = x[opEnd-9]
	_ = x[opBasicLit-10]
	_ = x[opStrictIntLit-11]
	_ = x[opStrictFloatLit-12]
	_ = x[opStrictCharLit-13]
	_ = x[opStrictStringLit-14]
	_ = x[opStrictComplexLit-15]
	_ = x[opIdent-16]
	_ = x[opStdlibPkg-17]
	_ = x[opIndexExpr-18]
	_ = x[opSliceExpr-19]
	_ = x[opSliceFromExpr-20]
	_ = x[opSliceToExpr-21]
	_ = x[opSliceFromToExpr-22]
	_ = x[opSliceToCapExpr-23]
	_ = x[opSliceFromToCapExpr-24]
	_ = x[opFuncLit-25]
	_ = x[opCompositeLit-26]
	_ = x[opTypedCompositeLit-27]
	_ = x[opSimpleSelectorExpr-28]
	_ = x[opSelectorExpr-29]
	_ = x[opTypeAssertExpr-30]
	_ = x[opTypeSwitchAssertExpr-31]
	_ = x[opVoidFuncType-32]
	_ = x[opFuncType-33]
	_ = x[opArrayType-34]
	_ = x[opSliceType-35]
	_ = x[opMapType-36]
	_ = x[opChanType-37]
	_ = x[opKeyValueExpr-38]
	_ = x[opEllipsis-39]
	_ = x[opTypedEllipsis-40]
	_ = x[opStarExpr-41]
	_ = x[opUnaryExpr-42]
	_ = x[opBinaryExpr-43]
	_ = x[opParenExpr-44]
	_ = x[opVariadicCallExpr-45]
	_ = x[opNonVariadicCallExpr-46]
	_ = x[opCallExpr-47]
	_ = x[opAssignStmt-48]
	_ = x[opMultiAssignStmt-49]
	_ = x[opBranchStmt-50]
	_ = x[opSimpleLabeledBranchStmt-51]
	_ = x[opLabeledBranchStmt-52]
	_ = x[opSimpleLabeledStmt-53]
	_ = x[opLabeledStmt-54]
	_ = x[opBlockStmt-55]
	_ = x[opExprStmt-56]
	_ = x[opGoStmt-57]
	_ = x[opDeferStmt-58]
	_ = x[opSendStmt-59]
	_ = x[opEmptyStmt-60]
	_ = x[opIncDecStmt-61]
	_ = x[opReturnStmt-62]
	_ = x[opIfStmt-63]
	_ = x[opIfInitStmt-64]
	_ = x[opIfElseStmt-65]
	_ = x[opIfInitElseStmt-66]
	_ = x[opIfNamedOptStmt-67]
	_ = x[opIfNamedOptElseStmt-68]
	_ = x[opSwitchStmt-69]
	_ = x[opSwitchTagStmt-70]
	_ = x[opSwitchInitStmt-71]
	_ = x[opSwitchInitTagStmt-72]
	_ = x[opSelectStmt-73]
	_ = x[opTypeSwitchStmt-74]
	_ = x[opTypeSwitchInitStmt-75]
	_ = x[opCaseClause-76]
	_ = x[opDefaultCaseClause-77]
	_ = x[opCommClause-78]
	_ = x[opDefaultCommClause-79]
	_ = x[opForStmt-80]
	_ = x[opForPostStmt-81]
	_ = x[opForCondStmt-82]
	_ = x[opForCondPostStmt-83]
	_ = x[opForInitStmt-84]
	_ = x[opForInitPostStmt-85]
	_ = x[opForInitCondStmt-86]
	_ = x[opForInitCondPostStmt-87]
	_ = x[opRangeStmt-88]
	_ = x[opRangeKeyStmt-89]
	_ = x[opRangeKeyValueStmt-90]
	_ = x[opFieldList-91]
	_ = x[opUnnamedField-92]
	_ = x[opSimpleField-93]
	_ = x[opField-94]
	_ = x[opMultiField-95]
	_ = x[opValueInitSpec-96]
	_ = x[opTypedValueInitSpec-97]
	_ = x[opTypedValueSpec-98]
	_ = x[opTypeSpec-99]
	_ = x[opTypeAliasSpec-100]
	_ = x[opFuncDecl-101]
	_ = x[opMethodDecl-102]
	_ = x[opFuncProtoDecl-103]
	_ = x[opMethodProtoDecl-104]
	_ = x[opConstDecl-105]
	_ = x[opVarDecl-106]
	_ = x[opTypeDecl-107]
	_ = x[opEmptyPackage-108]
}

const _operation_name = "InvalidNodeNamedNodeNodeSeqNamedNodeSeqOptNodeNamedOptNodeMultiStmtMultiExprEndBasicLitStrictIntLitStrictFloatLitStrictCharLitStrictStringLitStrictComplexLitIdentStdlibPkgIndexExprSliceExprSliceFromExprSliceToExprSliceFromToExprSliceToCapExprSliceFromToCapExprFuncLitCompositeLitTypedCompositeLitSimpleSelectorExprSelectorExprTypeAssertExprTypeSwitchAssertExprVoidFuncTypeFuncTypeArrayTypeSliceTypeMapTypeChanTypeKeyValueExprEllipsisTypedEllipsisStarExprUnaryExprBinaryExprParenExprVariadicCallExprNonVariadicCallExprCallExprAssignStmtMultiAssignStmtBranchStmtSimpleLabeledBranchStmtLabeledBranchStmtSimpleLabeledStmtLabeledStmtBlockStmtExprStmtGoStmtDeferStmtSendStmtEmptyStmtIncDecStmtReturnStmtIfStmtIfInitStmtIfElseStmtIfInitElseStmtIfNamedOptStmtIfNamedOptElseStmtSwitchStmtSwitchTagStmtSwitchInitStmtSwitchInitTagStmtSelectStmtTypeSwitchStmtTypeSwitchInitStmtCaseClauseDefaultCaseClauseCommClauseDefaultCommClauseForStmtForPostStmtForCondStmtForCondPostStmtForInitStmtForInitPostStmtForInitCondStmtForInitCondPostStmtRangeStmtRangeKeyStmtRangeKeyValueStmtFieldListUnnamedFieldSimpleFieldFieldMultiFieldValueInitSpecTypedValueInitSpecTypedValueSpecTypeSpecTypeAliasSpecFuncDeclMethodDeclFuncProtoDeclMethodProtoDeclConstDeclVarDeclTypeDeclEmptyPackage"

var _operation_index = [...]uint16{0, 7, 11, 20, 27, 39, 46, 58, 67, 76, 79, 87, 99, 113, 126, 141, 157, 162, 171, 180, 189, 202, 213, 228, 242, 260, 267, 279, 296, 314, 326, 340, 360, 372, 380, 389, 398, 405, 413, 425, 433, 446, 454, 463, 473, 482, 498, 517, 525, 535, 550, 560, 583, 600, 617, 628, 637, 645, 651, 660, 668, 677, 687, 697, 703, 713, 723, 737, 751, 769, 779, 792, 806, 823, 833, 847, 865, 875, 892, 902, 919, 926, 937, 948, 963, 974, 989, 1004, 1023, 1032, 1044, 1061, 1070, 1082, 1093, 1098, 1108, 1121, 1139, 1153, 1161, 1174, 1182, 1192, 1205, 1220, 1229, 1236, 1244, 1256}

func (i operation) String() string {
	if i >= operation(len(_operation_index)-1) {
		return "operation(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _operation_name[_operation_index[i]:_operation_index[i+1]]
}
