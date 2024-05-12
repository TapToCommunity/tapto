# Expected scanner behavior

Tapping a card on the reader will execute the card commands, or launch a game.

## With exit_game=no ( default )

If a game is launched with a card

- removing the card from the reader won't close the game
- leaving the card on the reader will have no effects
- tapping another game will launch another game without going in the core menu
- tapping the same card will load the game again from scratch
- tapping a command like insert coin will execute the command will not interrupt the game
- exiting the game manually with the internal menu and forget previous state so you will be able to tap any other card
- exiting the game manually with the card still on the reader won't relaunch the game once on the menu


## With exit_game=yes and exit_game_delay=0

If a game is launched with a card

- removing the card from the reader will close the game
- with thetapping another game will launch another game without going in the core menu

## With exit_game=yes and exit_game_delay=N

If a game is launched with a card

- removing the card from the reader will close the game after N seconds
- removing the card from the reader and reinserting it before N seconds will not interrupt the current game
- removing the card from the reader and tapping another game will launch the other game immediately
- removing the card from the reader and tapping another command will execute the command and reset the delay timer, you can tap X cards that will not change software and reset the timer each time, and reinsert the previous game and keep the session running


