package main

import (
	"bytes"
	"math/rand"

	//"encoding/base64"
	//"encoding/hex"
	"encoding/binary"
	"fmt"
	"io/ioutil"

	//"os"
	//"path/filepath"
	//"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("Rock Maze")
	//fmt.Println(fmt.Sprintf("%02x", math.Float32bits(8100)))
	data, err := ioutil.ReadFile("testquest.bin") //testquest
	if err != nil {
		panic(err)
	}

	// 0 = EMPTY
	// 1 = ROCK
	// 2 = END
	// 3 = START
	// 9 = OUT OF BOUNDS
	var maze = [10][10]int{
		{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
		{9, 1, 1, 1, 1, 9, 9, 9, 9, 9},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 9},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 9},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{9, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{9, 9, 1, 1, 1, 1, 1, 1, 1, 1},
		{9, 9, 9, 9, 9, 1, 1, 1, 1, 9},
	}

	var xoffset float32 = 13000 //12400
	var zoffset float32 = 12800 //12400
	var xscale int = 700        // 800
	var yscale int = 650

	var startposx int = 0
	var startposy int = 3
	var endposx int = 9
	var endposy int = 6

	maze = carveMaze(4+rand.Intn(3)-rand.Intn(3), 4+rand.Intn(3)-rand.Intn(3), maze, 0)

	if rand.Intn(2) == 1 { //Start on left or right
		startposx = 9
		startposy = 6
		endposx = 0
		endposy = 3
	}
	maze[endposy][endposx] = 2
	maze[startposy][startposx] = 3

	//Lazy solution to the rare unsolvable maze.
	if maze[endposy-1][endposx] == 1 && maze[endposy+1][endposx] == 1 {
		if endposx == 0 {
			if maze[endposy][endposx+1] == 1 {
				maze[endposy-1][endposx] = 0
			}
		} else {
			if maze[endposy][endposx-1] == 1 {
				maze[endposy+1][endposx] = 0
			}
		}
	}
	if maze[startposy-1][startposx] == 1 && maze[startposy+1][startposx] == 1 {
		if startposx == 0 {
			if maze[startposy][startposx+1] == 1 {
				maze[startposy-1][startposx] = 0
			}
		} else {
			if maze[startposy][startposx-1] == 1 {
				maze[startposy+1][startposx] = 0
			}
		}
	}

	//Target Transition
	//X
	data = writeFloat(data, xoffset-float32((xscale-70)*startposx), 1540)
	//Z
	data = writeFloat(data, zoffset-float32(yscale*startposy), 1548)

	//Red Box
	//X
	data = writeFloat(data, xoffset-float32(xscale*endposx), 3236)
	//Z
	data = writeFloat(data, zoffset-float32(yscale*endposy), 3244)

	//Powderstone
	//X
	data = writeFloat(data, xoffset-float32(xscale*startposx), 3264)
	//Z
	data = writeFloat(data, zoffset-float32(yscale*startposy), 3272)

	var size int = 10
	var rocknum int = 0

	//Memory Locations
	var monsteridloc int = 3396
	var monsterxloc int = 3428

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if rocknum > 38 {
				fmt.Print("rock limit exceeded")
				i = size
				j = size
				break
			}

			switch maze[i][j] { //Which object?
			case 0:
				fmt.Print("_")
				continue //Skips to next object
			case 1:
				fmt.Print("#")
				data[monsteridloc+(60*rocknum)] = 29  //Rock - Each table is 60 bytes
				data[monsteridloc+2+(60*rocknum)] = 7 // 7 is bigger, 6 is smaller but can break
				break
			case 2:
				fmt.Print(maze[i][j])
				data[monsteridloc+(60*rocknum)] = 86 //Cactus
				data[monsteridloc+2+(60*rocknum)] = 1
				data = writeFloat(data, 200, monsterxloc+4+(60*rocknum))
				break
			case 3:
				fmt.Print(maze[i][j])
				data[monsteridloc+(60*rocknum)] = 29
				data[monsteridloc+2+(60*rocknum)] = 2 //Gem Rock
				data = writeFloat(data, 200, monsterxloc+4+(60*rocknum))
				break
			case 9:
				fmt.Print("X")
				continue
			}
			if maze[i][j] == 1 || maze[i][j] == 2 || maze[i][j] == 3 {
				//Writing the object's coordinates
				//X
				data = writeFloat(data, xoffset-float32(xscale*j), monsterxloc+(60*rocknum))
				//Y
				data = writeFloat(data, 0, monsterxloc+4+(60*rocknum))
				//Z
				data = writeFloat(data, zoffset-float32(yscale*i), monsterxloc+8+(60*rocknum))
				rocknum++
			}

		}
		fmt.Println()
	}

	//Writes finished file
	err = ioutil.WriteFile("quest_override.bin", data, 0644)

	fmt.Println("data written")
}

func carveMaze(x int, y int, maze [10][10]int, depth int) [10][10]int {
	//This is kind of a mess but works. I dunno.
	depth++
	if depth == 30 { //Just in case.
		return maze
	}
	maze[x][y] = 0
	var carved bool = false
	switch rand.Intn(4) {
	case 0:
		if x >= 2 {
			if maze[x-2][y] == 1 {
				maze[x-2][y] = 0
				maze[x-1][y] = 0
				maze = carveMaze(x-2, y, maze, depth)
				carved = true
			}
		} /*else {
			if x == 1 {
				maze[0][y] = 0
				maze = carveMaze(0, y, maze)
				carved = true
			}
		}*/
		if carved {
			break
		}
	case 1:
		if x <= 7 {
			if maze[x+2][y] == 1 {
				maze[x+2][y] = 0
				maze[x+1][y] = 0
				maze = carveMaze(x+2, y, maze, depth)
				carved = true
			}
		} /*else {
			if x == 7 {
				maze[8][y] = 0
				maze = carveMaze(8, y, maze)
				carved = true
			}
		}*/
		if carved {
			break
		}
	case 2:
		if y >= 2 {
			if maze[x][y-2] == 1 {
				maze[x][y-2] = 0
				maze[x][y-1] = 0
				maze = carveMaze(x, y-2, maze, depth)
				carved = true
			}
		} /*else {
			if y == 1 {
				maze[x][1] = 0
				maze = carveMaze(x, 1, maze)
				carved = true
			}
		}*/
		if carved {
			break
		}
	case 3:
		if y <= 7 {
			if maze[x][y+2] == 1 {
				maze[x][y+2] = 0
				maze[x][y+1] = 0
				maze = carveMaze(x, y+2, maze, depth)
				carved = true
			} /*else {
				if y == 7 {
					maze[x][8] = 0
					maze = carveMaze(x, 8, maze)
					carved = true
				}
			}*/
		}
		if carved {
			break
		}
	}
	if x >= 2 {
		if maze[x-2][y] == 1 {
			maze[x-2][y] = 0
			maze[x-1][y] = 0
			maze = carveMaze(x-2, y, maze, depth)
		}
	}
	if x <= 7 {
		if maze[x+2][y] == 1 {
			maze[x+2][y] = 0
			maze[x+1][y] = 0
			maze = carveMaze(x+2, y, maze, depth)
		}
	}
	if y >= 2 {
		if maze[x][y-2] == 1 {
			maze[x][y-2] = 0
			maze[x][y-1] = 0
			maze = carveMaze(x, y-2, maze, depth)
		}
	}
	if y <= 7 {
		if maze[x][y+2] == 1 {
			maze[x][y+2] = 0
			maze[x][y+1] = 0
			maze = carveMaze(x, y+2, maze, depth)
		}
	}
	return maze
}

func writeFloat(buf []byte, in float32, index int) []byte {
	var fl []byte = float32ToByte(in)
	buf[index] = fl[0]
	buf[index+1] = fl[1]
	buf[index+2] = fl[2]
	buf[index+3] = fl[3]
	return buf
}

func float32ToByte(f float32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}
