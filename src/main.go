package main

type model struct {
	board board
}

type board struct {
	unsolved []tile
	selected []tile
}

type tile struct {
	word  string
	group int
}

func initialModel() model {
	return model{
		board: board{
			unsolved: []tile{
				{
					word:  "Red",
					group: 1,
				},
				{
					word:  "Green",
					group: 1,
				},
				{
					word:  "Blue",
					group: 1,
				},
				{
					word:  "Yellow",
					group: 1,
				},
				{
					word:  "Cat",
					group: 2,
				},
				{
					word:  "Dog",
					group: 2,
				},
				{
					word:  "Hamster",
					group: 2,
				},
				{
					word:  "Fish",
					group: 2,
				},
				{
					word:  "Ghana",
					group: 3,
				},
				{
					word:  "Namibia",
					group: 3,
				},
				{
					word:  "Morocco",
					group: 3,
				},
				{
					word:  "Ethiopia",
					group: 3,
				},
				{
					word:  "Mercedes",
					group: 4,
				},
				{
					word:  "Kia",
					group: 4,
				},
				{
					word:  "Toyota",
					group: 4,
				},
				{
					word:  "Ford",
					group: 4,
				},
			},
			selected: []tile{},
		},
	}
}
