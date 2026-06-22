package utils

import (
	"math"
)

type RoutePoint struct {
	Index     int
	Lng       float64
	Lat       float64
	Name      string
	Priority  int
	PointType int
}

type OptimizedRoute struct {
	OrderedPoints  []RoutePoint
	TotalDistance  float64
	DistanceMatrix [][]float64
	Strategy       int
	StrategyName   string
}

const EarthRadius = 6371000.0

func HaversineDistance(lng1, lat1, lng2, lat2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

func BuildDistanceMatrix(points []RoutePoint) [][]float64 {
	n := len(points)
	matrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i != j {
				matrix[i][j] = HaversineDistance(points[i].Lng, points[i].Lat, points[j].Lng, points[j].Lat)
			}
		}
	}
	return matrix
}

func calculateRouteDistance(startLng, startLat float64, ordered []RoutePoint) float64 {
	if len(ordered) == 0 {
		return 0
	}
	total := HaversineDistance(startLng, startLat, ordered[0].Lng, ordered[0].Lat)
	for i := 1; i < len(ordered); i++ {
		total += HaversineDistance(ordered[i-1].Lng, ordered[i-1].Lat, ordered[i].Lng, ordered[i].Lat)
	}
	return total
}

func twoOpt(startLng, startLat float64, route []RoutePoint) ([]RoutePoint, float64) {
	n := len(route)
	if n < 4 {
		return route, calculateRouteDistance(startLng, startLat, route)
	}
	improved := true
	bestDist := calculateRouteDistance(startLng, startLat, route)
	bestRoute := make([]RoutePoint, n)
	copy(bestRoute, route)
	for improved {
		improved = false
		for i := 0; i < n-1; i++ {
			for j := i + 1; j < n; j++ {
				newRoute := make([]RoutePoint, n)
				copy(newRoute[:i], bestRoute[:i])
				for k := 0; k <= j-i; k++ {
					newRoute[i+k] = bestRoute[j-k]
				}
				copy(newRoute[j+1:], bestRoute[j+1:])
				newDist := calculateRouteDistance(startLng, startLat, newRoute)
				if newDist < bestDist-1e-6 {
					bestDist = newDist
					copy(bestRoute, newRoute)
					improved = true
				}
			}
		}
	}
	return bestRoute, bestDist
}

func NearestNeighborTSP(startLng, startLat float64, points []RoutePoint, distanceMatrix [][]float64) OptimizedRoute {
	n := len(points)
	visited := make([]bool, n)
	ordered := make([]RoutePoint, 0, n)
	currentLng, currentLat := startLng, startLat

	for i := 0; i < n; i++ {
		nearestIdx := -1
		minDist := math.MaxFloat64
		for j := 0; j < n; j++ {
			if !visited[j] {
				d := HaversineDistance(currentLng, currentLat, points[j].Lng, points[j].Lat)
				if d < minDist {
					minDist = d
					nearestIdx = j
				}
			}
		}
		if nearestIdx >= 0 {
			visited[nearestIdx] = true
			ordered = append(ordered, points[nearestIdx])
			currentLng, currentLat = points[nearestIdx].Lng, points[nearestIdx].Lat
		}
	}

	ordered, totalDist := twoOpt(startLng, startLat, ordered)
	return OptimizedRoute{
		OrderedPoints:  ordered,
		TotalDistance:  totalDist,
		DistanceMatrix: distanceMatrix,
		Strategy:       10,
		StrategyName:   "速度优先(最近邻)",
	}
}

func GreedyInsertionTSP(startLng, startLat float64, points []RoutePoint, distanceMatrix [][]float64) OptimizedRoute {
	n := len(points)
	if n == 0 {
		return OptimizedRoute{OrderedPoints: []RoutePoint{}, TotalDistance: 0, DistanceMatrix: distanceMatrix, Strategy: 11, StrategyName: "距离最短(贪心插入)"}
	}
	if n == 1 {
		totalDist := HaversineDistance(startLng, startLat, points[0].Lng, points[0].Lat)
		return OptimizedRoute{OrderedPoints: points, TotalDistance: totalDist, DistanceMatrix: distanceMatrix, Strategy: 11, StrategyName: "距离最短(贪心插入)"}
	}

	unvisited := make(map[int]RoutePoint)
	for i, p := range points {
		unvisited[i] = p
	}

	firstIdx := -1
	firstDist := math.MaxFloat64
	for i, p := range points {
		d := HaversineDistance(startLng, startLat, p.Lng, p.Lat)
		if d < firstDist {
			firstDist = d
			firstIdx = i
		}
	}
	ordered := []RoutePoint{points[firstIdx]}
	delete(unvisited, firstIdx)

	for len(unvisited) > 0 {
		bestInsertIdx := -1
		bestPointIdx := -1
		minIncrease := math.MaxFloat64

		for pointIdx, point := range unvisited {
			for insertPos := 0; insertPos <= len(ordered); insertPos++ {
				var distBefore, distAfter, increase float64
				if insertPos == 0 {
					distBefore = HaversineDistance(startLng, startLat, point.Lng, point.Lat)
					distAfter = HaversineDistance(point.Lng, point.Lat, ordered[0].Lng, ordered[0].Lat)
					original := HaversineDistance(startLng, startLat, ordered[0].Lng, ordered[0].Lat)
					increase = distBefore + distAfter - original
				} else if insertPos == len(ordered) {
					distBefore = HaversineDistance(ordered[len(ordered)-1].Lng, ordered[len(ordered)-1].Lat, point.Lng, point.Lat)
					increase = distBefore
				} else {
					distBefore = HaversineDistance(ordered[insertPos-1].Lng, ordered[insertPos-1].Lat, point.Lng, point.Lat)
					distAfter = HaversineDistance(point.Lng, point.Lat, ordered[insertPos].Lng, ordered[insertPos].Lat)
					original := HaversineDistance(ordered[insertPos-1].Lng, ordered[insertPos-1].Lat, ordered[insertPos].Lng, ordered[insertPos].Lat)
					increase = distBefore + distAfter - original
				}
				if increase < minIncrease {
					minIncrease = increase
					bestInsertIdx = insertPos
					bestPointIdx = pointIdx
				}
			}
		}

		if bestPointIdx >= 0 {
			newOrdered := make([]RoutePoint, 0, len(ordered)+1)
			newOrdered = append(newOrdered, ordered[:bestInsertIdx]...)
			newOrdered = append(newOrdered, unvisited[bestPointIdx])
			newOrdered = append(newOrdered, ordered[bestInsertIdx:]...)
			ordered = newOrdered
			delete(unvisited, bestPointIdx)
		}
	}

	ordered, totalDist := twoOpt(startLng, startLat, ordered)
	return OptimizedRoute{
		OrderedPoints:  ordered,
		TotalDistance:  totalDist,
		DistanceMatrix: distanceMatrix,
		Strategy:       11,
		StrategyName:   "距离最短(贪心插入)",
	}
}

func InsertionOrderTSP(startLng, startLat float64, points []RoutePoint, distanceMatrix [][]float64) OptimizedRoute {
	n := len(points)
	if n == 0 {
		return OptimizedRoute{OrderedPoints: []RoutePoint{}, TotalDistance: 0, DistanceMatrix: distanceMatrix, Strategy: 12, StrategyName: "优先级优先"}
	}

	idxList := make([]int, n)
	for i := range idxList {
		idxList[i] = i
	}

	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			priDiff := points[idxList[j]].Priority - points[idxList[i]].Priority
			if priDiff < 0 || (priDiff == 0 && idxList[j] < idxList[i]) {
				idxList[i], idxList[j] = idxList[j], idxList[i]
			}
		}
	}

	ordered := make([]RoutePoint, n)
	for i, idx := range idxList {
		ordered[i] = points[idx]
	}

	_, totalDist := twoOpt(startLng, startLat, ordered)
	return OptimizedRoute{
		OrderedPoints:  ordered,
		TotalDistance:  totalDist,
		DistanceMatrix: distanceMatrix,
		Strategy:       12,
		StrategyName:   "优先级优先",
	}
}

func WeightedNearestNeighbor(startLng, startLat float64, points []RoutePoint, distanceMatrix [][]float64) OptimizedRoute {
	n := len(points)
	if n == 0 {
		return OptimizedRoute{OrderedPoints: []RoutePoint{}, TotalDistance: 0, DistanceMatrix: distanceMatrix, Strategy: 13, StrategyName: "综合最优(加权)"}
	}

	visited := make([]bool, n)
	ordered := make([]RoutePoint, 0, n)
	currentLng, currentLat := startLng, startLat

	distanceWeight := 0.5
	priorityWeight := 0.3
	typeWeight := 0.2

	var maxDist, minDist, maxPriority, minPriority, maxType, minType float64
	for i, p := range points {
		d := HaversineDistance(startLng, startLat, p.Lng, p.Lat)
		if i == 0 || d > maxDist {
			maxDist = d
		}
		if i == 0 || d < minDist {
			minDist = d
		}
		pri := float64(p.Priority)
		if pri > maxPriority {
			maxPriority = pri
		}
		if pri < minPriority {
			minPriority = pri
		}
		pt := float64(p.PointType)
		if pt > maxType {
			maxType = pt
		}
		if pt < minType {
			minType = pt
		}
	}

	normalize := func(v, minV, maxV float64) float64 {
		if maxV == minV {
			return 0.5
		}
		return (v - minV) / (maxV - minV)
	}

	for i := 0; i < n; i++ {
		bestIdx := -1
		minScore := math.MaxFloat64
		for j := 0; j < n; j++ {
			if !visited[j] {
				d := HaversineDistance(currentLng, currentLat, points[j].Lng, points[j].Lat)
				normDist := normalize(d, minDist, maxDist)
				normPri := normalize(float64(points[j].Priority), minPriority, maxPriority)
				normType := normalize(float64(points[j].PointType), minType, maxType)

				score := distanceWeight*normDist - priorityWeight*normPri - typeWeight*normType

				if score < minScore {
					minScore = score
					bestIdx = j
				}
			}
		}
		if bestIdx >= 0 {
			visited[bestIdx] = true
			ordered = append(ordered, points[bestIdx])
			currentLng, currentLat = points[bestIdx].Lng, points[bestIdx].Lat
		}
	}

	ordered, totalDist := twoOpt(startLng, startLat, ordered)
	return OptimizedRoute{
		OrderedPoints:  ordered,
		TotalDistance:  totalDist,
		DistanceMatrix: distanceMatrix,
		Strategy:       13,
		StrategyName:   "综合最优(加权)",
	}
}

func OptimizeRoute(startLng, startLat float64, points []RoutePoint, strategy int) OptimizedRoute {
	distanceMatrix := BuildDistanceMatrix(points)

	switch strategy {
	case 11:
		return GreedyInsertionTSP(startLng, startLat, points, distanceMatrix)
	case 12:
		return InsertionOrderTSP(startLng, startLat, points, distanceMatrix)
	case 13:
		return WeightedNearestNeighbor(startLng, startLat, points, distanceMatrix)
	case 10:
		fallthrough
	default:
		return NearestNeighborTSP(startLng, startLat, points, distanceMatrix)
	}
}
