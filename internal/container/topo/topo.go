package topo

import "fmt"

// DependencyEdge bean依赖边
type DependencyEdge struct {
    From string `json:"from"` // 依赖者
    To   string `json:"to"`   // 被依赖者
}

// detectCircularDependency 检测bean循环依赖
//
// Args:
// - edges: bean依赖关系列表
//
// Returns: 循环依赖路径，nil表示无循环依赖
func detectCircularDependency(edges []*DependencyEdge) []string {
    edgeMap, _ := buildDependencyGraph(edges)

    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    // 遍历所有bean节点
    for beanName := range edgeMap {
        if !visited[beanName] {
            if foundCycle, cycle := dfsDetectCycle(beanName, edgeMap, visited, recStack, nil); foundCycle {
                return cycle
            }
        }
    }

    return nil
}

// Sort 拓扑排序获取bean创建顺序
//
// Args:
// - edges: bean依赖关系列表
//
// Returns: bean创建顺序，错误信息
func Sort(edges []*DependencyEdge) ([]string, error) {
    edgeMap, inDegree := buildDependencyGraph(edges)

    // 检查循环依赖
    if cycle := detectCircularDependency(edges); cycle != nil {
        return nil, fmt.Errorf("circular dependency detected: %v", cycle)
    }

    // Kahn算法拓扑排序
    queue := getZeroInDegreeNodes(inDegree)
    result := make([]string, 0)

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        result = append(result, current)

        // 减少邻接节点的入度
        for _, edge := range edgeMap[current] {
            inDegree[edge.To]--
            if inDegree[edge.To] == 0 {
                queue = append(queue, edge.To)
            }
        }
    }

    return result, nil
}

// 构建bean依赖图和入度表
//
// Args:
// - edges: bean依赖关系列表
//
// Returns: 邻接表，入度表
func buildDependencyGraph(edges []*DependencyEdge) (map[string][]*DependencyEdge, map[string]int) {
    edgeMap := make(map[string][]*DependencyEdge)
    inDegree := make(map[string]int)

    // 初始化所有节点
    for _, edge := range edges {
        if _, exists := inDegree[edge.From]; !exists {
            inDegree[edge.From] = 0
        }
        if _, exists := inDegree[edge.To]; !exists {
            inDegree[edge.To] = 0
        }
    }

    // 构建邻接表和入度
    for _, edge := range edges {
        edgeMap[edge.From] = append(edgeMap[edge.From], edge)
        inDegree[edge.To]++
    }

    return edgeMap, inDegree
}

// 获取入度为0的节点（无依赖的bean）
func getZeroInDegreeNodes(inDegree map[string]int) []string {
    var nodes []string
    for node, degree := range inDegree {
        if degree == 0 {
            nodes = append(nodes, node)
        }
    }
    return nodes
}

// 检测循环依赖
func dfsDetectCycle(node string, edgeMap map[string][]*DependencyEdge, visited map[string]bool, recStack map[string]bool, currentPath []string) (bool, []string) {
    if recStack[node] {
        // 找到循环，构造循环路径
        for i, n := range currentPath {
            if n == node {
                cycle := make([]string, len(currentPath[i:]))
                copy(cycle, currentPath[i:])
                cycle = append(cycle, node) // 闭合循环
                return true, cycle
            }
        }
    }
    
    if visited[node] {
        return false, nil
    }

    visited[node] = true
    recStack[node] = true
    currentPath = append(currentPath, node)

    // 遍历所有依赖
    for _, edge := range edgeMap[node] {
        if foundCycle, cycle := dfsDetectCycle(edge.To, edgeMap, visited, recStack, currentPath); foundCycle {
            return true, cycle
        }
    }

    recStack[node] = false
    return false, nil
}