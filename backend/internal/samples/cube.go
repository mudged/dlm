package samples

import (
	"example.com/dlm/backend/internal/wiremodel"
)

// corner returns vertex i of axis-aligned cube [−1,1]³ (edge length 2 m).
func corner(i int) [3]float64 {
	xs := []float64{-1, 1, 1, -1, -1, 1, 1, -1}
	ys := []float64{-1, -1, 1, 1, -1, -1, 1, 1}
	zs := []float64{-1, -1, -1, -1, 1, 1, 1, 1}
	return [3]float64{xs[i] * cubeHalf, ys[i] * cubeHalf, zs[i] * cubeHalf}
}

// undirectedEdges lists the 12 cube edges as corner indices.
var undirectedEdges = [][2]int{
	{0, 1}, {1, 2}, {2, 3}, {3, 0},
	{4, 5}, {5, 6}, {6, 7}, {7, 4},
	{0, 4}, {1, 5}, {2, 6}, {3, 7},
}

// CubeLights walks each undirected edge twice (both directions) in an Eulerian
// circuit so every edge is covered; consecutive Euclidean spacing is chordTarget (REQ-009).
func CubeLights() []wiremodel.Light {
	adj := make([][]int, 8)
	for _, e := range undirectedEdges {
		u, v := e[0], e[1]
		adj[u] = append(adj[u], v)
		adj[v] = append(adj[v], u)
	}
	verts := eulerianVertexPath(copyAdj(adj), 0)
	var pts [][3]float64
	for i := 0; i < len(verts)-1; i++ {
		a := corner(verts[i])
		b := corner(verts[i+1])
		seg := subdivideEdge(a[0], a[1], a[2], b[0], b[1], b[2])
		pts = appendChain(pts, seg)
	}
	return assignIDs(pts)
}

func copyAdj(adj [][]int) [][]int {
	out := make([][]int, len(adj))
	for i := range adj {
		out[i] = append([]int(nil), adj[i]...)
	}
	return out
}

// eulerianVertexPath returns a vertex sequence v0,v1,…,v0 for a directed multigraph
// where each undirected edge appears as u→v and v→u (24 directed edges).
func eulerianVertexPath(adj [][]int, start int) []int {
	var stack []int
	var circuit []int
	stack = append(stack, start)
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		if len(adj[u]) == 0 {
			stack = stack[:len(stack)-1]
			circuit = append(circuit, u)
			continue
		}
		v := adj[u][len(adj[u])-1]
		adj[u] = adj[u][:len(adj[u])-1]
		stack = append(stack, v)
	}
	for i, j := 0, len(circuit)-1; i < j; i, j = i+1, j-1 {
		circuit[i], circuit[j] = circuit[j], circuit[i]
	}
	return circuit
}

// NameCube is the canonical English sample name (REQ-009).
const NameCube = "Sample cube"
