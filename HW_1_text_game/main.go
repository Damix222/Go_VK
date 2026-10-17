package main

import (
	"fmt"
	"slices"
	"strings"
)

type Game struct {
	player Player
	rooms  map[string]*Room
	door   Door
}

type Player struct {
	currentRoom  *Room
	inventory    []Item
	backpackIsOn bool
}

type Item struct {
	name     string
	location string
}

type Room struct {
	name             string
	itemsInRoom      []Item
	otherRooms       map[string]*Room
	orderedRooms     []string
	orderedLocations []string
	description      string
}

type Door struct {
	isOpen bool
}

var game Game

func (game *Game) changeCurrentRoom(roomName string) string {
	nextRoom, ok := game.player.currentRoom.otherRooms[roomName]
	if !ok {
		return fmt.Sprintf("нет пути в %s", roomName)
	}

	if nextRoom.name != "улица" {
		game.player.currentRoom = nextRoom
		return nextRoom.roomDescription()
	} else if game.door.isOpen {
		game.player.currentRoom = nextRoom
		return nextRoom.roomDescription()
	}

	return "дверь закрыта"
}

func (room Room) roomDescription() string {
	if room.name == "кухня" {
		return fmt.Sprintf("кухня, ничего интересного. можно пройти - %s",
			strings.Join(room.orderedRooms, ", "),
		)
	}
	return fmt.Sprintf("%s. можно пройти - %s",
		room.description,
		strings.Join(room.orderedRooms, ", "),
	)
}

func (room Room) showItemsInRoom() string {
	items := make(map[string][]string)

	for _, item := range room.itemsInRoom {
		items[item.location] = append(items[item.location], item.name)
	}

	finalStr := []string{}

	for _, location := range room.orderedLocations {
		if len(items[location]) == 0 {
			continue
		}

		part := fmt.Sprintf("%s: %s",
			location,
			strings.Join(items[location], ", "),
		)

		finalStr = append(finalStr, part)
	}

	return strings.Join(finalStr, ", ")
}

func (room Room) lookAround() string {
	if room.name == "кухня" {
		return room.showKitchen()
	} else {
		s := room.showItemsInRoom()
		if s == "" {
			s = "пустая комната"
		}

		finalStr := s + ". можно пройти - " + strings.Join(room.orderedRooms, ", ")

		return finalStr
	}
}

func (room Room) showKitchen() string {
	if !(game.player.backpackIsOn) || !game.player.isItemIn("ключи") || !game.player.isItemIn("конспекты") {
		return fmt.Sprintf("%s, надо собрать рюкзак и идти в универ. можно пройти - %s",
			room.description,
			strings.Join(room.orderedRooms, ", "),
		)
	}

	return fmt.Sprintf("%s, надо идти в универ. можно пройти - %s",
		room.description,
		strings.Join(room.orderedRooms, ", "),
	)
}

func (room *Room) wearBackpack() string {
	for i, v := range room.itemsInRoom {
		if v.name == "рюкзак" {
			room.itemsInRoom = slices.Delete(room.itemsInRoom, i, i+1)
			game.player.backpackIsOn = true
			return "вы надели: рюкзак"
		}
	}

	return "рюкзака нет в этой комнате"
}

func (room *Room) takeItem(itemName string) string {
	for i, item := range room.itemsInRoom {
		if item.name == itemName {

			if !game.player.backpackIsOn {
				return "некуда класть"
			}

			room.itemsInRoom = slices.Delete(room.itemsInRoom, i, i+1)
			game.player.inventory = append(game.player.inventory, item)
			return fmt.Sprintf("предмет добавлен в инвентарь: %s", itemName)
		}
	}

	return "нет такого"
}

func (player Player) useItem(item1, item2 string) string {
	for _, item := range player.inventory {
		if item.name == item1 {
			if item1 == "ключи" && item2 == "дверь" {
				game.door.isOpen = true
				return "дверь открыта"
			}

			return "не к чему применить"
		}
	}

	return fmt.Sprintf("нет предмета в инвентаре - %s", item1)
}

func (player Player) isItemIn(itemName string) bool {
	for _, item := range player.inventory {
		if item.name == itemName {
			return true
		}
	}

	return false
}

func initGame() {
	// ПРЕДМЕТЫ
	keys := Item{
		name:     "ключи",
		location: "на столе",
	}

	backpack := Item{
		name:     "рюкзак",
		location: "на стуле",
	}

	notes := Item{
		name:     "конспекты",
		location: "на столе",
	}

	// СОЗДАНИЕ КОМНАТ
	room := Room{
		name:             "комната",
		itemsInRoom:      []Item{keys, backpack, notes},
		orderedLocations: []string{"на столе", "на стуле"},
		description:      "ты в своей комнате",
	}

	kitchen := Room{
		name:        "кухня",
		description: "ты находишься на кухне, на столе: чай",
	}

	hall := Room{
		name:        "коридор",
		description: "ничего интересного",
	}

	street := Room{
		name:        "улица",
		description: "на улице весна",
	}

	// КУДА МОЖНО ПРОЙТИ
	hall.otherRooms = map[string]*Room{
		room.name:    &room,
		kitchen.name: &kitchen,
		street.name:  &street,
	}

	hall.orderedRooms = []string{
		kitchen.name,
		room.name,
		street.name,
	}

	room.otherRooms = map[string]*Room{
		hall.name: &hall,
	}

	room.orderedRooms = []string{
		hall.name,
	}

	kitchen.otherRooms = map[string]*Room{
		hall.name: &hall,
	}

	kitchen.orderedRooms = []string{
		hall.name,
	}

	street.otherRooms = map[string]*Room{
		"домой": &hall,
	}

	street.orderedRooms = []string{
		"домой",
	}

	player1 := Player{
		currentRoom: &kitchen,
	}

	door := Door{
		isOpen: false,
	}

	game = Game{
		player: player1,
		rooms: map[string]*Room{
			room.name:    &room,
			kitchen.name: &kitchen,
			street.name:  &street,
			hall.name:    &hall,
		},

		door: door,
	}

}

func handleCommand(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "ошибка"
	}

	cmd := fields[0]

	switch cmd {
	case "идти":
		if len(fields) == 2 {
			return game.changeCurrentRoom(fields[1])
		}
		return "ошибка"

	case "осмотреться":
		return game.player.currentRoom.lookAround()

	case "надеть":
		if len(fields) == 2 && fields[1] == "рюкзак" {
			return game.player.currentRoom.wearBackpack()
		}
		return "неизвестная команда"

	case "взять":
		if len(fields) == 2 {
			return game.player.currentRoom.takeItem(fields[1])
		}
		return "неизвестная команда"

	case "применить":
		if len(fields) == 3 {
			return game.player.useItem(fields[1], fields[2])
		}
		return "неизвестная команда"

	default:
		return "неизвестная команда"
	}
}

func main() {
	initGame()

	fmt.Println(handleCommand("осмотреться"))
}
