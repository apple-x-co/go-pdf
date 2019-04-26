package types

const WrapContent = "wrap_content"
const MatchParent = "match_parent"

type LayoutSize string

func (L LayoutSize) IsWrapContent() bool {
	return L == WrapContent || L == ""
}

func (L LayoutSize) IsMatchParent() bool {
	return L == MatchParent
}
