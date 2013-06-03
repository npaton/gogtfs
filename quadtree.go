// Package quadtree implements methods for a quadtree spatial partitioning data
// structure.
//
// Code is based on the Wikipedia article
// http://en.wikipedia.org/wiki/Quadtree.
//
// Adapted from github.com/foolusion/quadtree
//
package gtfs

import (
	"math"
)

// node_capacity is the maximum number of points allowed in a quadtree node
var node_capacity int = 4

// AABB represents an Axis-Aligned bounding box structure with center and half
// dimension
type AABB struct {
	centerX  float64
	centerY  float64
	halfDimX float64
	halfDimY float64
}

// NewAABB creates a new axis-aligned bounding box and returns its address
func NewAABB(centerX, centerY, halfDimX, halfDimY float64) *AABB {
	return &AABB{centerX, centerY, halfDimX, halfDimY}
}

// ContainsPoint returns true when the AABB contains the point given
func (aabb *AABB) ContainsPoint(p *Stop) bool {
	if p.Lon < aabb.centerX-aabb.halfDimX {
		return false
	}
	if p.Lat < aabb.centerY-aabb.halfDimY {
		return false
	}
	if p.Lon > aabb.centerX+aabb.halfDimX {
		return false
	}
	if p.Lat > aabb.centerY+aabb.halfDimY {
		return false
	}

	return true
}

// IntersectsAABB returns true when the AABB intersects another AABB
func (aabb *AABB) IntersectsAABB(other *AABB) bool {
	if other.centerX+other.halfDimX < aabb.centerX-aabb.halfDimX {
		return false
	}
	if other.centerY+other.halfDimY < aabb.centerY-aabb.halfDimY {
		return false
	}
	if other.centerX-other.halfDimX > aabb.centerX+aabb.halfDimX {
		return false
	}
	if other.centerY-other.halfDimY > aabb.centerY+aabb.halfDimY {
		return false
	}

	return true
}

// QuadTree represents the quadtree data structure.
type QuadTree struct {
	boundary  AABB
	points    []*Stop
	northWest *QuadTree
	northEast *QuadTree
	southWest *QuadTree
	southEast *QuadTree
}

// New creates a new quadtree node that is bounded by boundary and contains
// node_capacity points.
func CreateQuadtree(minLat, maxLat, minLon, maxLon float64) *QuadTree {
	halfdimX := (maxLon - minLon) / 2
	halfdimY := (maxLat - minLat) / 2
	centerX := halfdimX + minLon
	centerY := halfdimY + minLat
	boundary := *NewAABB(centerX, centerY, halfdimX, halfdimY)
	return NewQuadtree(boundary)
}

func NewQuadtree(boundary AABB) *QuadTree {
	points := make([]*Stop, 0, node_capacity)
	qt := &QuadTree{boundary: boundary, points: points}
	return qt
}

// Insert adds a point to the quadtree. It returns true if it was successful
// and false otherwise.
func (qt *QuadTree) Insert(p *Stop) bool {
	// Ignore objects which do not belong in this quad tree.
	if !qt.boundary.ContainsPoint(p) {
		return false
	}

	// If there is space in this quad tree, add the object here.
	if len(qt.points) < cap(qt.points) {
		qt.points = append(qt.points, p)
		return true
	}

	// Otherwise, we need to subdivide then add the point to whichever node
	// will accept it.
	if qt.northWest == nil {
		qt.subDivide()
	}

	if qt.northWest.Insert(p) {
		return true
	}
	if qt.northEast.Insert(p) {
		return true
	}
	if qt.southWest.Insert(p) {
		return true
	}
	if qt.southEast.Insert(p) {
		return true
	}

	// Otherwise, the point cannot be inserted for some unknown reason.
	// (which should never happen)
	return false
}

func (qt *QuadTree) subDivide() {
	// Check if this is a leaf node.
	if qt.northWest != nil {
		return
	}

	box := AABB{
		qt.boundary.centerX - qt.boundary.halfDimX/2, qt.boundary.centerY + qt.boundary.halfDimY/2,
		qt.boundary.halfDimX / 2, qt.boundary.halfDimY / 2}
	qt.northWest = NewQuadtree(box)

	box = AABB{
		qt.boundary.centerX + qt.boundary.halfDimX/2, qt.boundary.centerY + qt.boundary.halfDimY/2,
		qt.boundary.halfDimX / 2, qt.boundary.halfDimY / 2}
	qt.northEast = NewQuadtree(box)

	box = AABB{
		qt.boundary.centerX - qt.boundary.halfDimX/2, qt.boundary.centerY - qt.boundary.halfDimY/2,
		qt.boundary.halfDimX / 2, qt.boundary.halfDimY / 2}
	qt.southWest = NewQuadtree(box)

	box = AABB{
		qt.boundary.centerX + qt.boundary.halfDimX/2, qt.boundary.centerY - qt.boundary.halfDimY/2,
		qt.boundary.halfDimX / 2, qt.boundary.halfDimY / 2}
	qt.southEast = NewQuadtree(box)

	for _, v := range qt.points {
		if qt.northWest.Insert(v) {
			continue
		}
		if qt.northEast.Insert(v) {
			continue
		}
		if qt.southWest.Insert(v) {
			continue
		}
		if qt.southEast.Insert(v) {
			continue
		}
	}
	qt.points = nil
}

func (qt *QuadTree) SearchByProximity(lat, lng, radius float64) (results []*Stop) {
	earth_radius := 6371000.0 // in meter
	x1 := lng - (180.0 / math.Pi * (radius / earth_radius / math.Cos(lat*math.Pi/180.0)))
	x2 := lng + (180.0 / math.Pi * (radius / earth_radius / math.Cos(lat*math.Pi/180.0)))
	y1 := lat + (radius / earth_radius * 180.0 / math.Pi)
	y2 := lat - (radius / earth_radius * 180.0 / math.Pi)
	// FIXME: this is bounding box search, not radial
	return qt.SearchArea(NewAABB(lng, lat, math.Abs(x1-x2)/2.0, math.Abs(y1-y2)/2.0))
}

func (qt *QuadTree) SearchArea(a *AABB) []*Stop {
	results := make([]*Stop, 0, node_capacity)

	if !qt.boundary.IntersectsAABB(a) {
		return results
	}

	for _, v := range qt.points {
		if a.ContainsPoint(v) {
			results = append(results, v)
		}
	}

	if qt.northWest == nil {
		return results
	}

	results = append(results, qt.northWest.SearchArea(a)...)
	results = append(results, qt.northEast.SearchArea(a)...)
	results = append(results, qt.southWest.SearchArea(a)...)
	results = append(results, qt.southEast.SearchArea(a)...)
	return results
}
