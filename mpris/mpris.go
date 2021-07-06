package mpris

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

const (
	mprisIface = "org.mpris.MediaPlayer2"
	mprisPath  = "/org/mpris/MediaPlayer2"

	mprisPrefix = mprisIface + "."
)

// Player provides access to a particular player instance.
type Player struct {
	Name string
	conn *dbus.Conn
	obj  *utils.DBusObject
}

// Players returns all of the players on a particular bus.
func Players(conn *dbus.Conn) ([]*Player, error) {
	names, err := utils.DBusListNames(conn.BusObject())
	if err != nil {
		return nil, fmt.Errorf("failed to list object names: %w", err)
	}

	var players []*Player

	for _, name := range names {
		if strings.HasPrefix(name, mprisPrefix) {
			players = append(players, &Player{
				Name: name,
				conn: conn,
				obj:  utils.NewDBusObject(conn, name, mprisPath),
			})
		}
	}

	return players, nil
}

// SubscribePlayers subscribes to changes in players.
func SubscribePlayers(conn *dbus.Conn) (<-chan utils.NameOwnerChange, utils.UnsubFunc, error) {
	changeChan, unsub, err := utils.DBusSubscribeNameOwnerChanged(conn)

	if err != nil {
		return nil, nil, err
	}

	filteredChangeChan := make(chan utils.NameOwnerChange, 1)

	go func() {
		defer close(filteredChangeChan)

		for change := range changeChan {
			if !strings.HasPrefix(change.Name, mprisPrefix) {
				continue
			}

			filteredChangeChan <- change
		}
	}()

	return filteredChangeChan, unsub, nil
}

// SubscribePlaybackStatus subscribes to changes in the current playback
// status.
func (p *Player) SubscribePropertyChanges() (<-chan utils.PropertiesChange, utils.UnsubFunc, error) {
	changeChan, unsub, err := utils.DBusSubscribePropertyChanges(p.conn)
	if err != nil {
		return nil, nil, err
	}

	filteredChangeChan := make(chan utils.PropertiesChange, 1)

	go func() {
		defer close(filteredChangeChan)

		for change := range changeChan {
			if change.Signal.Sender != p.Name {
				continue
			}

			filteredChangeChan <- change
		}
	}()

	return filteredChangeChan, unsub, nil
}
