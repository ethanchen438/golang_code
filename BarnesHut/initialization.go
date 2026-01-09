package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// InitializeUniverse() sets an initial universe given a collection of galaxies and a width.
// It returns a pointer to the resulting universe.
func InitializeUniverse(galaxies []Galaxy, w float64) *Universe {
	var u Universe
	u.width = w
	u.stars = make([]*Star, 0, len(galaxies)*len(galaxies[0]))
	for i := range galaxies {
		for _, b := range galaxies[i] {
			u.stars = append(u.stars, b)
		}
	}
	return &u
}

// InitializeGalaxy takes number of stars in the galaxy, radius of the galaxy to be constructed,
// and center of galaxy to be constructed. Returns a spinning Galaxy object -- which is just a slice of Star pointers
func InitializeGalaxy(numOfStars int, r, x, y float64) Galaxy {
	g := make(Galaxy, numOfStars)

	for i := range g {
		var s Star

		// First choose distance to center of galaxy
		dist := (rand.Float64() + 1.0) / 2.0

		// multiply by factor of r
		dist *= r

		// Next choose the angle in radians to represent the rotation
		angle := rand.Float64() * 2 * math.Pi

		// convert polar coordinates to Cartesian
		s.position.x = x + dist*math.Cos(angle)
		s.position.y = y + dist*math.Sin(angle)

		// set the mass = mass of sun by default
		s.mass = solarMass

		// set the radius equal to radius of sun in m
		s.radius = 696340000

		//set the colors
		s.red = 255
		s.green = 255
		s.blue = 255

		// now spin the galaxy

		// the following is orbital velocity equation
		//dist := Distance(pos, g[i].position)
		speed := 0.5 * math.Sqrt(G*blackHoleMass/dist) // approximation of orbital velocity equation: half of true speed to prevent instability

		s.velocity.x = speed * math.Cos(angle+math.Pi/2.0)
		s.velocity.y = speed * math.Sin(angle+math.Pi/2.0)

		//point g[i] at s
		g[i] = &s

	}

	//add a blackhole to the center of the galaxy

	var blackhole Star
	blackhole.mass = blackHoleMass
	blackhole.position.x = x
	blackhole.position.y = y
	blackhole.blue = 255
	blackhole.radius = 6963400000 // ten times that of a normal star (to make it visible as large)

	g = append(g, &blackhole)

	return g
}

// ParseOrderedPair remains the same as its functionality is correct for the new OrderedPair type.
func ParseOrderedPair(line string) (OrderedPair, error) {
	// Replace the Unicode minus sign with a standard hyphen-minus
	line = strings.ReplaceAll(line, "−", "-")

	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		return OrderedPair{}, fmt.Errorf("invalid ordered pair")
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return OrderedPair{}, err
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return OrderedPair{}, err
	}
	return OrderedPair{x: x, y: y}, nil
}

// ParseRGB remains the same as its functionality is correct for RGB values.
func ParseRGB(line string) (uint8, uint8, uint8, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid RGB format")
	}
	red, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, 0, err
	}
	green, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, 0, err
	}
	blue, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return 0, 0, 0, err
	}
	return uint8(red), uint8(green), uint8(blue), nil
}

func ReadUniverse(filename string) (Universe, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Universe{}, err
	}
	defer file.Close()

	var universe Universe
	scanner := bufio.NewScanner(file)

	// Read the first line to get the width of the universe
	if scanner.Scan() {
		width, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64)
		if err != nil {
			return Universe{}, fmt.Errorf("invalid universe width: %v", err)
		}
		universe.width = width
	} else {
		return Universe{}, fmt.Errorf("file is empty or missing width")
	}

	// === CHANGE 1: Skip the line for gravitational constant (G) ===
	// G is now a package-level constant (6.67408e-11) and should not be read from the file.
	if !scanner.Scan() {
		// Should have a G value, even if we are ignoring it
		return Universe{}, fmt.Errorf("file is missing gravitational constant line")
	}

	var currentStar *Star // Use a pointer to Star
	lineType := 0         // Keeps track of which data is expected next

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch lineType {
		case 0: // Expecting a star name, starting with '>'
			if strings.HasPrefix(line, ">") {
				// If there was a previous star, add it to the universe's stars slice
				if currentStar != nil {
					// === CHANGE 2: Append pointer to Star to stars slice ===
					universe.stars = append(universe.stars, currentStar)
				}
				// Start a new star, initializing a pointer to a new Star object
				currentStar = &Star{}
				// The body name (e.g., ">Jupiter") is not stored in the new Star struct,
				// so we only proceed to the next expected line type.
				lineType = 1
			} else {
				return Universe{}, fmt.Errorf("expected star name, got: %s", line)
			}

		case 1: // Expecting RGB values
			red, green, blue, err := ParseRGB(line)
			if err != nil {
				return Universe{}, fmt.Errorf("invalid RGB values: %v", err)
			}
			currentStar.red, currentStar.green, currentStar.blue = red, green, blue
			lineType = 2

		case 2: // Expecting mass
			mass, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return Universe{}, fmt.Errorf("invalid mass: %v", err)
			}
			currentStar.mass = mass
			lineType = 3

		case 3: // Expecting radius
			radius, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return Universe{}, fmt.Errorf("invalid radius: %v", err)
			}
			currentStar.radius = radius
			lineType = 4

		case 4: // Expecting position (OrderedPair)
			position, err := ParseOrderedPair(line)
			if err != nil {
				return Universe{}, fmt.Errorf("invalid position: %v", err)
			}
			currentStar.position = position
			lineType = 5

		case 5: // Expecting velocity (OrderedPair)
			velocity, err := ParseOrderedPair(line)
			if err != nil {
				return Universe{}, fmt.Errorf("invalid velocity: %v", err)
			}
			currentStar.velocity = velocity
			// The Star struct has an acceleration field, but the input file doesn't provide it,
			// so it remains the zero value (x: 0.0, y: 0.0)
			lineType = 0 // Ready for the next star
		}
	}

	// Add the last star, if there is one
	if currentStar != nil {
		universe.stars = append(universe.stars, currentStar)
	}

	if err := scanner.Err(); err != nil {
		return Universe{}, err
	}

	return universe, nil
}

