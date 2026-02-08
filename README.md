# gun-shooter

Result: https://nobonobo.github.io/gun-shooter/

# scenes
- top
  - create room -> room
- room(start listen)
  - start game -> game
  - QR code(join with hostID) -> join with hostID
- join with hostID -> scope
  - nickname(form)
- game
  - quit -> room
- scope (connect to hostID)
  - quit -> room

# data types

- pointer position (client -> host)
  - nickname
  - x, y (float)
  - trigger (boolean)
