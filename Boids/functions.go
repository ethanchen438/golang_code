// Name: Ethan Chen
// Date: 09/29/2025

package main

import "math"

// Input: the current sky and boid number
// Output: an updated acceleration based off three different parameters, separation, alignment, and cohesion
func UpdateAcceleration(currentSky Sky, i int) OrderedPair {

	var newAcceleration OrderedPair
	count := 0

	for j := range currentSky.boids {
		// Checks to make sure we are not updating based off the same boid
		if i != j {
			dis := ComputeDistance(currentSky.boids[i], currentSky.boids[j])
			// Checks to make sure the boids are not in the exact same position
			if dis == 0 {
				continue
			}
			// Checks to make sure the distance between the boids are not too far from each other to compute a reasonable acceleration
			if dis >= currentSky.proximity {
				continue
			}
			// Checks to make sure the distance between the boids is within range before calculating forces
			if dis < currentSky.proximity {
				sep := ComputeSeparation(currentSky.boids[i], currentSky.boids[j], currentSky.separationFactor, dis)
				align := ComputeAlignment(currentSky.boids[j], currentSky.alignmentFactor, dis)
				cohesion := ComputeCohesion(currentSky.boids[i], currentSky.boids[j], currentSky.cohesionFactor, dis)
				newAcceleration.x += sep.x + align.x + cohesion.x
				newAcceleration.y += sep.y + align.y + cohesion.y
				count++
			}
		}
	}
	// Ensures we have more than one relationship between two boids before normalizing the accelerations
	if count > 0 {
		newAcceleration.x /= float64(count)
		newAcceleration.y /= float64(count)
	}
	return newAcceleration
}

// Input: two boids, the separation factor and distance between them
// Output: Calculates the separation force using an equation taking into account all the inputs
func ComputeSeparation(b, b2 Boid, separationFactor, distance float64) OrderedPair {

	var separationForce OrderedPair

	if distance == 0 {
		return separationForce
	}

	separationForce.x = (b.position.x - b2.position.x) / (distance * distance) * separationFactor
	separationForce.y = (b.position.y - b2.position.y) / (distance * distance) * separationFactor

	return separationForce
}

// Input: current boid, the alignment factor and distance between them
// Output: Calculates the alignment force using an equation taking into account all the inputs
func ComputeAlignment(b Boid, alignmentFactor, distance float64) OrderedPair {

	var alignmentForce OrderedPair

	if distance == 0 {
		return alignmentForce
	}

	alignmentForce.x = alignmentFactor * (b.velocity.x / distance)
	alignmentForce.y = alignmentFactor * (b.velocity.y / distance)

	return alignmentForce
}

// Input: two boids, the cohesion factor and distance between them
// Output: Calculates the cohesion force using an equation taking into account all the inputs
func ComputeCohesion(b, b2 Boid, cohesionFactor, distance float64) OrderedPair {

	var cohesionForce OrderedPair

	if distance == 0 {
		return cohesionForce
	}

	cohesionForce.x = cohesionFactor * (b2.position.x - b.position.x) / distance
	cohesionForce.y = cohesionFactor * (b2.position.y - b.position.y) / distance

	return cohesionForce
}

// Input: two boids
// Output: the squared distance between the two boids
func ComputeDistance(b, b2 Boid) float64 {

	disx := b.position.x - b2.position.x
	disy := b.position.y - b2.position.y

	return math.Sqrt(disx*disx + disy*disy)
}

// Input: a boid, old acceleration, maxboidspeed, and a timestep
// Output: an updated velocity calculated off a formula taking into account all the inputs
func UpdateVelocity(b Boid, oldAcceleration OrderedPair, maxBoidSpeed, timeStep float64) OrderedPair {

	var newVelocity OrderedPair

	newVelocity.x = 0.5*(b.acceleration.x+oldAcceleration.x)*timeStep + b.velocity.x
	newVelocity.y = 0.5*(b.acceleration.y+oldAcceleration.y)*timeStep + b.velocity.y

	speed := math.Sqrt(newVelocity.x*newVelocity.x + newVelocity.y*newVelocity.y)
	// Ensures that the velocity has a cap and does not exceed the maxboidspeed while also not being 0
	if speed > maxBoidSpeed && speed > 0 {
		newVelocity.x = (newVelocity.x / speed) * maxBoidSpeed
		newVelocity.y = (newVelocity.y / speed) * maxBoidSpeed
	}
	return newVelocity
}

// Input: a boid, old acceleration and velocity, skywidth, and a timestep
// Output: an updated position calculated based off a formula taking into account all the inputs
func UpdatePosition(b Boid, oldAcceleration, oldVelocity OrderedPair, skyWidth, timeStep float64) OrderedPair {

	var newPosition OrderedPair

	newPosition.x = 0.5*oldAcceleration.x*(timeStep*timeStep) + oldVelocity.x*timeStep + b.position.x
	newPosition.y = 0.5*oldAcceleration.y*(timeStep*timeStep) + oldVelocity.y*timeStep + b.position.y

	// Checks for wrap around when a boid reaches the edge of the canvas
	for newPosition.x < 0 {
		newPosition.x += skyWidth
	}
	for newPosition.x > skyWidth {
		newPosition.x -= skyWidth
	}
	for newPosition.y < 0 {
		newPosition.y += skyWidth
	}
	for newPosition.y > skyWidth {
		newPosition.y -= skyWidth
	}
	return newPosition
}

// Input: a sky and timeStep
// Output: a new updated sky based on the timestep value passed
func UpdateSky(currentSky Sky, timeStep float64) Sky {

	newSky := copySky(currentSky)

	for i, b := range newSky.boids {

		oldAcceleration := b.acceleration
		oldVelocity := b.velocity
		newSky.boids[i].acceleration = UpdateAcceleration(currentSky, i)
		newSky.boids[i].velocity = UpdateVelocity(newSky.boids[i], oldAcceleration, newSky.maxBoidSpeed, timeStep)
		newSky.boids[i].position = UpdatePosition(newSky.boids[i], oldAcceleration, oldVelocity, newSky.width, timeStep)
	}
	return newSky
}

// Input: a sky
// Output: a sky copy of the input sky
func copySky(currentSky Sky) Sky {

	var newSky Sky
	newSky.width = currentSky.width
	newSky.proximity = currentSky.proximity
	newSky.alignmentFactor = currentSky.alignmentFactor
	newSky.cohesionFactor = currentSky.cohesionFactor
	newSky.separationFactor = currentSky.separationFactor
	newSky.maxBoidSpeed = currentSky.maxBoidSpeed
	numBoids := len(currentSky.boids)
	newSky.boids = make([]Boid, numBoids)

	for i := range newSky.boids {
		newSky.boids[i] = copyBoid(currentSky.boids[i])
	}
	return newSky
}

// Input: a boid
// Output: a boid copy of the input boid
func copyBoid(b Boid) Boid {

	var b2 Boid
	b2.velocity.x = b.velocity.x
	b2.velocity.y = b.velocity.y
	b2.acceleration.x = b.acceleration.x
	b2.acceleration.y = b.acceleration.y
	b2.position.x = b.position.x
	b2.position.y = b.position.y

	return b2
}

// Input: an initial Sky, a number of generations, and a timestep interval.
// Return: a slice of Skies of length numGens+1 to simulate the Boids model over numGens generations, using the initial Sky.
func SimulateBoids(initialSky Sky, numGens int, timeStep float64) []Sky {

	timepoints := make([]Sky, numGens+1)
	timepoints[0] = initialSky

	for i := 1; i < numGens+1; i++ {
		timepoints[i] = UpdateSky(timepoints[i-1], timeStep)
	}
	return timepoints
}
