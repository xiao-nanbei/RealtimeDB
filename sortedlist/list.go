package sortedlist

import "math"

// List 实现了排序链表的数据结构
type List interface {
	// Remove 移除节点
	Remove(key int64) bool

	// Add 新增节点
	Add(key int64, data interface{})

	// Range 过滤范围内的 key 并返回 Iter 对象
	Range(lower, upper int64) Iter

	// All 迭代所有对象
	All() Iter

	PreAll() Iter
}

// Iter 迭代器对象
type Iter interface {
	Pre() bool
	Next() bool
	Value() interface{}
}

type iter struct {
	cursor int
	data   []interface{}
}


// Next 推进迭代器
func (it *iter) Next() bool {
	it.cursor++
	if len(it.data) > it.cursor {
		return true
	}

	return false
}
// Next 推进迭代器
func (it *iter) Pre() bool {
	it.cursor--
	if it.cursor>=0 {
		return true
	}
	return false
}
// Value 返回迭代器当前 value
func (it *iter) Value() interface{} {
	return it.data[it.cursor]
}

const (
	RED bool = true
	BLACK bool = false
)
type Node struct {
	key   int64
	value interface{}
	left  *Node
	right *Node
	color bool
}

type redblackTree struct {
	root *Node
	size int
}
func NewNode(key int64, val interface{}) *Node {
	// 默认添加红节点
	return &Node{
		key:    key,
		value:  val,
		left:   nil,
		right:  nil,
		//parent: nil,
		color:  RED,
	}
}

// NewTree 生成 AVL 树
func NewTree() List {
	return &redblackTree{}
}
func (node *Node) IsRed() bool {
	if node == nil {
		return BLACK
	}
	return node.color
}
func (tree *redblackTree) GetTreeSize() int {
	return tree.size
}
func (tree *redblackTree) Add(k int64, v interface{}) {
	isAdd:=0
	isAdd,tree.root = tree.root.insert(k, v)
	tree.size += isAdd
	tree.root.color = BLACK //根节点为黑色节点
}

func (tree *redblackTree) Remove(k int64) bool {
	if tree.root.search(k) {
		tree.root.delete(k)
		return true
	}
	return false
}

func (tree *redblackTree) All() Iter {
	return tree.root.values(0, math.MaxInt64)
}
func (tree *redblackTree) PreAll() Iter{
	return tree.root.prevalues(0, math.MaxInt64)
}
func (tree *redblackTree) Range(lower, upper int64) Iter {
	return tree.root.values(lower, upper)
}

func (tree *redblackTree) getMaxValue() int64 {
	return tree.root.maxNode().key
}

func (tree *redblackTree) getMinValue() int64 {
	return tree.root.minNode().key
}

func (node *Node)insert(k int64, v interface{}) (int, *Node) {
	if node == nil { // 默认插入红色节点
		return 1,NewNode(k, v)
	}
	isAdd := 0
	if k < node.key {
		isAdd, node.left = node.left.insert(k, v)
	} else if k > node.key {
		isAdd, node.right = node.right.insert(k, v)
	} else {
		// 对value值更新,节点数量不增加,isAdd = 0
		node.value = v
	}

	// 维护红黑树
	node = node.updateRedBlackTree(isAdd)

	return isAdd, node
}

func (node *Node) search(k int64) bool {
	if node == nil {
		return false
	}
	cmp := k - node.key
	if cmp > 0 {
		// 如果 v 大于当前节点值，继续从右子树中寻找
		return node.right.search(k)
	} else if cmp < 0 {
		// 如果 v 小于当前节点值，继续从左子树中寻找
		return node.left.search(k)
	} else {
		// 相等则表示找到
		return true
	}
}

func (node *Node) delete(k int64) (int, *Node) {
	if node == nil {
		return 0,node
	}
	cmp := k - node.key
	isdelete:=0
	if cmp > 0 {
		// 如果 v 大于当前节点值，继续从右子树中删除
		isdelete,node.right = node.right.delete(k)
	} else if cmp < 0 {
		// 如果 v 小于当前节点值，继续从左子树中删除
		isdelete,node.left = node.left.delete(k)
	} else {
		// 找到 v
		if node.left != nil && node.right != nil {
			// 如果该节点既有左子树又有右子树
			// 使用右子树中的最小节点取代删除节点，然后删除右子树中的最小节点

			minnode := node.right.minNode()
			node.key = minnode.key
			node.value = minnode.value
			isdelete,node.right = node.right.delete(node.key)
		} else if node.left != nil {
			// 如果只有左子树，则直接删除节点
			node = node.left
		} else {
			// 只有右子树或空树
			node = node.right
		}
	}
	node = node.updateRedBlackTree(isdelete)
	return isdelete,node
}

func (node *Node) minNode() *Node {
	if node == nil {
		return nil
	}
	// 整棵树的最左边节点就是值最小的节点
	if node.left == nil {
		return node
	} else {
		return node.left.minNode()
	}
}

func (node *Node) maxNode() *Node {
	if node == nil {
		return nil
	}
	// 整棵树的最右边节点就是值最大的节点
	if node.right == nil {
		return node
	} else {
		return node.right.maxNode()
	}
}

func (node *Node) leftRotate() *Node {
	// 左旋转
	retNode := node.right
	node.right = retNode.left

	retNode.left = node
	retNode.color = node.color
	node.color = RED

	return retNode
}

//     node                    x
//    /   \     右旋转       /  \
//   x    T2   ------->   y   node
//  / \                       /  \
// y  T1                     T1  T2
func (node *Node) rightRotate() *Node {
	//右旋转
	retNode := node.left
	node.left = retNode.right

	retNode.right = node
	retNode.color = node.color
	node.color = RED

	return retNode
}
func (node *Node) flipColors() {
	node.color = RED
	node.left.color = BLACK
	node.right.color = BLACK
}
func (node *Node) updateRedBlackTree(isAdd int) *Node {
	// isAdd=0 说明没有新节点，无需维护
	if isAdd == 0 {
		return node
	}

	// 需要维护
	if node.right.IsRed() == RED && node.left.IsRed() != RED {
		node = node.leftRotate()
	}

	// 判断是否为情形3，是需要右旋转
	if node.left.IsRed() == RED && node.left.left.IsRed() == RED {
		node = node.rightRotate()
	}

	// 判断是否为情形4，是需要颜色翻转
	if node.left.IsRed() == RED && node.right.IsRed() == RED {
		node.flipColors()
	}

	return node

}


// appendValue 中序遍历按顺序获取所有值
func appendValue(values []interface{}, lower, upper int64, t *Node) []interface{} {
	if t != nil {
		values = appendValue(values, lower, upper, t.left)
		if t.key >= lower && t.key <= upper {
			values = append(values, t.value)
		}
		values = appendValue(values, lower, upper, t.right)
	}
	return values
}

func (node *Node) values(lower, upper int64) Iter {
	it := &iter{data: []interface{}{nil}}
	it.data = appendValue(it.data, lower, upper, node)

	return it
}
func (node *Node) prevalues(lower, upper int64) Iter {
	it := &iter{data: []interface{}{nil}}
	it.data = appendValue(it.data, lower, upper, node)
	it.cursor=len(it.data)-1
	return it
}

