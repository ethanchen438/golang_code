// Name: Ethan Chen
// Date: 10/26/25

package main

import "math"

// Input: an initial Universe, a number of generations, a timestep interval and a theta value
// Return: a slice of universes of length numGens+1 to simulate the BarnesHut model over numGens generations, using the initial Universe and theta value
func BarnesHut(initialUniverse *Universe, numGens int, time, theta float64) []*Universe {

	timePoints := make([]*Universe, numGens+1)
	timePoints[0] = initialUniverse

	for i := 1; i < numGens+1; i++ {
		timePoints[i] = updateUniverse(timePoints[i-1], time, theta)
	}
	return timePoints
}

// Input: a Universe, timestep and a theta value 
// Output: a new updated Universe based on the timestep and theta value passed
func updateUniverse(currentUniverse *Universe, time, theta float64) *Universe {

	newUniverse := copyUniverse(currentUniverse)
	tree := GenerateQuadTree(currentUniverse)

	for i, s := range newUniverse.stars {

		oldAcceleration := s.acceleration
		oldVelocity := s.velocity
		netForce := CalculateNetForce(tree.root, s, theta)
		newAcceleration := OrderedPair{
			x: netForce.x / s.mass,
			y: netForce.y / s.mass,
		}
		newUniverse.stars[i].acceleration = newAcceleration
		newUniverse.stars[i].velocity = UpdateVelocity(*s, oldAcceleration, time)
		newUniverse.stars[i].position = UpdatePosition(*s, oldAcceleration, oldVelocity, time)
	}
	return newUniverse
}

// Input: Takes in a universe
// Output: A universe copy of the universe passed
func copyUniverse(currentUniverse *Universe) *Universe {

	newUniverse := &Universe{
		width: currentUniverse.width,
		stars: make([]*Star, len(currentUniverse.stars)),
	}

	for i := range newUniverse.stars {
		newUniverse.stars[i] = copyStars(currentUniverse.stars[i])
	}
	return newUniverse
}

// Input: Takes in a star
// Output: A star copy of the star passed 
func copyStars(s *Star) *Star {

	s2 := &Star{}
	s2.velocity.x = s.velocity.x
	s2.velocity.y = s.velocity.y
	s2.position.x = s.position.x
	s2.position.y = s.position.y
	s2.acceleration.x = s.acceleration.x
	s2.acceleration.y = s.acceleration.y
	s2.mass = s.mass
	s2.radius = s.radius
	s2.red = s.red
	s2.green = s.green
	s2.blue = s.blue

	return s2
}

// Input: Takes in a universe 
// Output: Creates a quadtree based off all the stars in the universe 
func GenerateQuadTree(currentUniverse *Universe) *QuadTree {

	t := &QuadTree{}
	t.root = &Node{sector: Quadrant{x: 0, y: 0, width: currentUniverse.width}}

	for _, star := range currentUniverse.stars {
		t.root.insertStar(star)
	}
	return t
}

// Input: Takes in a star from a node
// Output: Recursively inserts the star into its proper quadrant of the quadtree
func (node *Node) insertStar(star *Star) {
	//Base Case
	if node.star == nil && node.children == nil {
		node.star = star
		return
	}
	// Children exist so do not need to be initialized 
	if node.children != nil {
		node.updateCenterOfMass(star)
		node.insertIntoChild(star)
		return
	}
	// Both stars occupy the same space 
	if node.star.position.x == star.position.x && node.star.position.y == star.position.y {
		node.star.mass = star.mass
		return
	}
	// Node has an existing star that needs to be recursively passed into the right quadrant 
	existing := node.star
	node.star = nil
	node.initializeQuadrant()
	node.insertIntoChild(existing)
	node.insertIntoChild(star)
	node.calculateCenterOfMass()
}

// Input: Takes in a star from a node
// Output: Recursive call to insert the star into the correct child 
func (node *Node) insertIntoChild(star *Star) {

	for _, child := range node.children {
		if inRange(star.position, child.sector) {
			child.insertStar(star)
			return
		}
	}
}

// Input: Takes in a node value 
// Output: Initializes the quadrant spaces for the node 
func (node *Node) initializeQuadrant() {

	q := node.sector
	half := q.width / 2
	node.children = make([]*Node, 4)
	node.children[0] = &Node{sector: Quadrant{x: q.x, y: q.y + half, width: half}}        // NW
	node.children[1] = &Node{sector: Quadrant{x: q.x + half, y: q.y + half, width: half}} // NE
	node.children[2] = &Node{sector: Quadrant{x: q.x, y: q.y, width: half}}               // SW
	node.children[3] = &Node{sector: Quadrant{x: q.x + half, y: q.y, width: half}}        // SE
}

// Input: Takes in a star from a node 
// Output: An updated center of mass based on the stars positions and their masses 
func (node *Node) updateCenterOfMass(newStar *Star) {
	// Initializes dummy star 
	if node.star == nil {
		node.star = &Star{position: newStar.position, mass: newStar.mass}
		return
	}
	totalMass := node.star.mass + newStar.mass
	node.star.position.x = (node.star.position.x*node.star.mass + newStar.position.x*newStar.mass) / totalMass
	node.star.position.y = (node.star.position.y*node.star.mass + newStar.position.y*newStar.mass) / totalMass
	node.star.mass = totalMass
}

// Input: Takes in a node value
// Output: The center of mass for one node based on all of their children and their relative position and mass
func (node *Node) calculateCenterOfMass() {
	// No children means no center of mass to be computed for BarnesHut
	if node.children == nil {
		return
	}
	var totalMass, xPos, yPos float64

	for _, child := range node.children {
		if child != nil && child.star != nil {
			totalMass += child.star.mass
			xPos += child.star.position.x * child.star.mass
			yPos += child.star.position.y * child.star.mass
		}
	}
	if totalMass > 0 {
		node.star = &Star{position: OrderedPair{x: xPos / totalMass, y: yPos / totalMass}, mass: totalMass}
	}
}

// Input: Takes in a position and a quadrant
// Output: a boolean value that checks if it is in range of the quadrant 
func inRange(pos OrderedPair, q Quadrant) bool {
	return pos.x >= q.x && pos.x < q.x+q.width && pos.y >= q.y && pos.y < q.y+q.width
}

// Input: Takes in a node, a star, and a theta value
// Output: The total net force acting on the star in the node from the current star 
func CalculateNetForce(node *Node, currStar *Star, theta float64) OrderedPair {

	var force OrderedPair
	s := node.sector.width
	dis := computeDistance(node.star, currStar)

	if dis == 0 || node.star == currStar {
		return force
	}

	ratio := s / dis

	if ratio < theta || node.children == nil {
		gForce := computeGravitationalForce(node.star, currStar, dis)
		force.x += gForce.x
		force.y += gForce.y
	}

	if ratio >= theta {
		for _, child := range node.children {
			if child.star != nil {
				childForce := CalculateNetForce(child, currStar, theta)
				force.x += childForce.x
				force.y += childForce.y
			}
		}
	}
	return force
}

// Input: Takes in two stars and the distance between them 
// Output: Using a gravitational constant, calculates the gravitational force between the two stars 
func computeGravitationalForce(s, s2 *Star, dis float64) OrderedPair {

	var gForce OrderedPair

	if dis == 0 {
		return gForce
	}
	gMag := G * s.mass * s2.mass / (dis * dis)
	dx := s.position.x - s2.position.x
	dy := s.position.y - s2.position.y
	gForce.x = gMag * dx / dis
	gForce.y = gMag * dy / dis

	return gForce
}

// Input: two stars
// Output: the squared distance between the two stars
func computeDistance(s, s2 *Star) float64 {

	disx := s2.position.x - s.position.x
	disy := s2.position.y - s.position.y

	return math.Sqrt(disx*disx + disy*disy)
}

// Input: a star, old acceleration, and a timestep
// Output: an updated velocity calculated off a formula taking into account all the inputs
func UpdateVelocity(s Star, oldAcceleration OrderedPair, timeStep float64) OrderedPair {

	var newVelocity OrderedPair
	newVelocity.x = 0.5*(s.acceleration.x+oldAcceleration.x)*timeStep + s.velocity.x
	newVelocity.y = 0.5*(s.acceleration.y+oldAcceleration.y)*timeStep + s.velocity.y

	return newVelocity
}

// Input: a star, old acceleration and velocity, and a timestep
// Output: an updated position calculated based off a formula taking into account all the inputs
func UpdatePosition(s Star, oldAcceleration, oldVelocity OrderedPair, timeStep float64) OrderedPair {

	var newPosition OrderedPair
	newPosition.x = 0.5*oldAcceleration.x*(timeStep*timeStep) + oldVelocity.x*timeStep + s.position.x
	newPosition.y = 0.5*oldAcceleration.y*(timeStep*timeStep) + oldVelocity.y*timeStep + s.position.y

	return newPosition
}