const WebSocket = require('ws');




const linh  = () => {

// Replace with your server URL
const serverUrl = 'ws://localhost:4296/game?gameID=12345';
const socket = new WebSocket(serverUrl);

socket.on('open', () => {
  console.log('Connected to WebSocket server');

  // Send a simple message after connecting (can simulate a game action)
  socket.send('Game Action: Player 1 started the game');
});

socket.on('message', (data) => {
  // Handle the game state sent by the server
  const gameState = JSON.parse(data);
  console.log(`Game state updated: ${gameState.Status}`);
});

socket.on('close', () => {
  console.log('Disconnected from WebSocket server');
});

socket.on('error', (error) => {
  console.error('WebSocket error:', error);
});
}

for (let i = 0; i < 100; i++) {
  linh()
}