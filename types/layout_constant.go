package types

const WrapContent = "wrap_content"
const MatchParent = "match_parent"

type LayoutConstant string

func (L LayoutConstant) IsWrapContent() bool {
	return L == WrapContent || L == ""
}

func (L LayoutConstant) IsMatchParent() bool {
	return L == MatchParent
}
