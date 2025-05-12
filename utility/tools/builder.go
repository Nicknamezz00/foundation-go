package tools

func BuildTree[T any](
	items []T,
	getID func(T) string,
	getParentID func(T) *string,
	getChildren func(*T) []*T,
	setChildren func(*T, []*T),
) []*T {
	idMap := make(map[string]*T)
	var roots []*T

	// 构建 ID 到节点的映射
	for i := range items {
		item := &items[i]
		idMap[getID(*item)] = item
	}

	// 构建树结构
	for i := range items {
		item := &items[i]
		parentID := getParentID(*item)
		if parentID == nil || *parentID == "" {
			roots = append(roots, item)
		} else if parent, ok := idMap[*parentID]; ok {
			children := getChildren(parent)
			children = append(children, item)
			setChildren(parent, children)
		}
	}

	return roots
}
