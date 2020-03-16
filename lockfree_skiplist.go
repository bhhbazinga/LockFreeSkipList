package lockfreeskiplist

import (
	"math/rand"
	"sync/atomic"
	"unsafe"
)

const maxLevel = 20

// LockFreeSkipList define
type LockFreeSkipList struct {
	head *node
	tail *node
	size int32
	comp func(value1 interface{}, value2 interface{}) bool
}

type node struct {
	level int
	nexts []unsafe.Pointer
	value interface{}
}

// NewLockFreeSkipList new a lockfree skiplist, you should pass a compare function.
func NewLockFreeSkipList(comp func(value1 interface{}, value2 interface{}) bool) *LockFreeSkipList {
	sl := new(LockFreeSkipList)
	sl.head = newNode(maxLevel, nil)
	sl.tail = newNode(maxLevel, nil)
	sl.size = 0
	sl.comp = comp
	for level := 0; level < maxLevel; level++ {
		sl.head.storeNext(level, sl.tail)
	}
	return sl
}

// Add a value to skiplist.
func (sl *LockFreeSkipList) Add(value interface{}) bool {
	var prevs [maxLevel]*node
	var nexts [maxLevel]*node
	for true {
		if sl.find(value, &prevs, &nexts) {
			return false
		}
		topLevel := randomLevel()
		newNode := newNode(topLevel, value)
		for level := 0; level < topLevel; level++ {
			newNode.storeNext(level, nexts[level])
		}
		prev := prevs[0]
		next := nexts[0]
		if !prev.casNext(0, next, newNode) {
			// The successor of prev is not next, we should try again.
			continue
		}
		for level := 1; level < topLevel; level++ {
			for true {
				prev := prevs[level]
				next := nexts[level]
				if prev.casNext(level, next, newNode) {
					break
				}
				// The successor of prev is not next,
				// we should call find to update the prevs and nexts.
				sl.find(value, &prevs, &nexts)
			}
		}
		break
	}
	atomic.AddInt32(&sl.size, 1)
	return true
}

// Remove a value from skiplist.
func (sl *LockFreeSkipList) Remove(value interface{}) bool {
	var prevs [maxLevel]*node
	var nexts [maxLevel]*node
	if !sl.find(value, &prevs, &nexts) {
		return false
	}
	removeNode := nexts[0]
	for level := removeNode.level - 1; level > 0; level-- {
		next := removeNode.loadNext(level)
		for !isMarked(next) {
			// Make sure that all but the bottom next are marked from top to bottom.
			removeNode.casNext(level, next, getMarked(next))
			next = removeNode.loadNext(level)
		}
	}
	next := removeNode.loadNext(0)
	for true {
		if isMarked(next) {
			// Other thread already maked the next, so this thread delete failed.
			return false
		}
		if removeNode.casNext(0, next, getMarked(next)) {
			// This thread marked the bottom next, delete successfully.
			break
		}
		next = removeNode.loadNext(0)
	}
	atomic.AddInt32(&sl.size, -1)
	return true
}

// Contains check if skiplist contains a value.
func (sl *LockFreeSkipList) Contains(value interface{}) bool {
	var prevs [maxLevel]*node
	var nexts [maxLevel]*node
	return sl.find(value, &prevs, &nexts)
}

// GetSize get the element size of skiplist.
func (sl *LockFreeSkipList) GetSize() int32 {
	return atomic.LoadInt32(&sl.size)
}

func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Int()&1 == 0 {
		level++
	}
	return level
}

func (sl *LockFreeSkipList) less(nd *node, value interface{}) bool {
	if sl.head == nd {
		return true
	}
	if sl.tail == nd {
		return false
	}
	return sl.comp(nd.value, value)
}

func (sl *LockFreeSkipList) equals(nd *node, value interface{}) bool {
	if sl.head == nd || sl.tail == nd {
		return false
	}
	return !sl.comp(nd.value, value) && !sl.comp(value, nd.value)
}

func (sl *LockFreeSkipList) find(value interface{}, prevs *[maxLevel]*node, nexts *[maxLevel]*node) bool {
	var prev *node
	var cur *node
	var next *node
retry:
	prev = sl.head
	for level := maxLevel - 1; level >= 0; level-- {
		cur = getUnmarked(prev.loadNext(level))
		for true {
			next = cur.loadNext(level)
			for isMarked(next) {
				// Like harris-linkedlist,remove the node while traversing.
				// See also https://github.com/bhhbazinga/LockFreeLinkedList.
				if !prev.casNext(level, cur, getUnmarked(next)) {
					goto retry
				}
				cur = getUnmarked(prev.loadNext(level))
				next = cur.loadNext(level)
			}
			if !sl.less(cur, value) {
				break
			}
			prev = cur
			cur = next
		}
		prevs[level] = prev
		nexts[level] = cur
	}
	return sl.equals(cur, value)
}

func newNode(level int, value interface{}) *node {
	nd := new(node)
	nd.nexts = make([]unsafe.Pointer, level)
	nd.level = level
	nd.value = value
	for level := 0; level < nd.level; level++ {
		nd.storeNext(level, nil)
	}
	return nd
}

func (nd *node) loadNext(level int) *node {
	return (*node)(atomic.LoadPointer(&nd.nexts[level]))
}

func (nd *node) storeNext(level int, next *node) {
	atomic.StorePointer(&nd.nexts[level], unsafe.Pointer(next))
}

func (nd *node) casNext(level int, expected *node, desire *node) bool {
	return atomic.CompareAndSwapPointer(&nd.nexts[level], unsafe.Pointer(expected), unsafe.Pointer(desire))
}

func isMarked(next *node) bool {
	ptr := unsafe.Pointer(next)
	return uintptr(ptr)&uintptr(0x1) == 1
}

func getMarked(next *node) *node {
	ptr := unsafe.Pointer(next)
	return (*node)(unsafe.Pointer(uintptr(ptr) | uintptr(0x1)))

}

func getUnmarked(next *node) *node {
	ptr := unsafe.Pointer(next)
	return (*node)(unsafe.Pointer(uintptr(ptr) & ^uintptr(0x1)))
}
