//URL management
const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
const userPicture = urlParams.get("picture");
const userPictureToSend = encodeURIComponent(userPicture); //profile picture binary to use to connect to websocket backend

//game DOM
document.getElementById("preGameLobby").style.backgroundImage = `url(data:image/jpeg;base64,${gamePicture})`;
let gameTimer = initialGameTimer;
let countdown;
const timerDisplay = document.getElementById("gameTimer");
const actionButton = document.getElementById("actionButton");
const scoreboard = document.getElementById("scoreboard");

//profile picture
if (gameAdmin) {
  const profilePicture = "{{.Account.ProfilePicture}}";
  if (profilePicture == "" || profilePicture == null) {
    document.getElementById("profilePicture").innerHTML = "<i class='bi bi-person-circle'></i>";
  }
  document.getElementById("profilePicture").src = `data:image/jpeg;base64,${profilePicture}`;
} else {
  document.getElementById("profilePicture").src = `data:image/jpeg;base64,${userPicture}`;
}


function startGameTimer() {
  // Display the initial time
  gameTimer = initialGameTimer;
  timerDisplay.textContent = gameTimer;

  countdown = setInterval(() => {
    if (gameTimer > 0) {
      gameTimer -= 1;
      timerDisplay.textContent = gameTimer;
    }
    if (gameTimer <= 0 && gameAdmin) {
      //ultimately, game admin should be the node addressing the time
      // Stop the timer when it reaches 0 and handle out-of-time scenarios
      clearInterval(countdown);
      ws.send(JSON.stringify(createClientMessage("outOfTime")));
      return;
    }
  }, 1000);
}
const userName = urlParams.get("name");
document.getElementById("userName").textContent = userName;
const createClientMessage = (label, message) => ({ label, message });

function action() {
  if (actionButton.innerText.toLowerCase() == "start") {
    actionButton.classList = "hidden bg-blue-400 p-5 rounded-xl absolute top-20 right-2 animate-jump-in animate-ease-in";
    // document.getElementById("preGameLobby").classList.add("hidden");
    document.getElementById("gameContent").classList.remove("hidden");
    ws.send(JSON.stringify(createClientMessage("startGame", "")));
  } else if (actionButton.innerText.toLowerCase() == "continue") {
    gameContent.classList.remove("hidden");
    gameContent.classList.add("block");
    scoreboard.classList.add("hidden");
    ws.send(JSON.stringify(createClientMessage("question", "")));
  }
  // actionButton.classList = "block bg-blue-400 p-5 rounded-xl absolute top-20 right-2 animate-jump-in animate-ease-in";
  // ws.send(JSON.stringify(createClientMessage("question", "")))
}
const ws = new WebSocket(`wss://${window.location.hostname}:${window.location.port}/play/${gameCode}/ws?name=${userName}&picture=${userPictureToSend}`);
ws.addEventListener("open", function (event) {
  showNotification("Connected to the game");
  console.log("picture used to connect is:", userPicture);
  const urlWithoutParams = window.location.href.split("?")[0];
  history.pushState({}, "", urlWithoutParams);
});
ws.addEventListener("message", function (event) {
  console.log(event.data);
  let rawData = JSON.parse(event.data);
  switch (String(rawData.label).trim()) {
    case "connectedUsers":
      if (!gameAdmin) {
        return;
      }
      message = JSON.parse(rawData.message);
      const playerListContainer = document.getElementById("player-list");
      if (gameAdmin && message.length >= 2) {
        const actionButton = document.getElementById("actionButton");
        actionButton.classList = "block bg-blue-400 p-5 rounded-xl absolute top-20 right-2 animate-jump-in animate-ease-in";
        // startButton.classList.remove("hidden")
        // startButton.classList.add("block")
        // startButton.classList.add("animate-jump-in")
        // startButton.classList.add("animate-ease-in")
      }

      // Iterate over the JSON data and create elements
      playerListContainer.innerHTML = "";
      message.forEach((player) => {
        if (player.name == userName) {
          return;
        }
        // Create a div for each player
        const playerDiv = document.createElement("div");
        playerDiv.className = "flex items-center mb-2"; // Apply Tailwind classes

        // Create an image element for the profile picture
        const profilePic = document.createElement("img");
        profilePic.src = `data:image/jpeg;base64,${player.picture}`;
        profilePic.alt = "PFP";
        profilePic.className = "h-10 w-10 rounded-full mr-0"; // Apply Tailwind classes

        // Create a span for the player's name
        const playerName = document.createElement("span");
        playerName.textContent = player.name;
        playerName.className = "text-lg"; // Apply Tailwind classes

        // Append the profile picture and name to the player div
        playerDiv.appendChild(profilePic);
        playerDiv.appendChild(playerName);
        playerDiv.classList.add("hover:underline");
        playerDiv.classList.add("mr-5");
        playerDiv.classList.add("bg-slate-400");
        playerDiv.classList.add("p-4");
        playerDiv.classList.add("rounded-lg");
        playerDiv.addEventListener("click", () => {
          const playerText = playerDiv.childNodes[1].innerText;
          ws.send(
            JSON.stringify({
              label: "kick",
              message: playerText,
            })
          );
          playerDiv.remove();
        });

        playerListContainer.appendChild(playerDiv);
      });
      return;
    case "leave":
      if (gameAdmin) {
        return;
      }
      location.replace(rawData.message);
      return;
    case "left":
      showNotification(rawData.message)
      return
    case "question":
      clearInterval(countdown);
      displayQuestion(JSON.parse(rawData.message));
      scoreboard.classList.add("hidden");
      return;
    case "scoreboard":
      gameTimer = 0;
      timerDisplay.innerText = 0;
      populateScoreboard(JSON.parse(rawData.message));
      return;
    case "endgame":
      question.innerText = "GAME OVER";
      ws.send(JSON.stringify(createClientMessage("scoreboard")));
      scoreboard.querySelector("h2").innerText = "Final Score for " + gameName;
      actionButton.classList.add("hidden");
      populateScoreboard(JSON.parse(rawData.message), true);
      return;
  }
});
ws.addEventListener("error", function (event) {
  showNotification("WebSocket error:", event);
});
const gameContent = document.getElementById("gameContent");
const question = document.getElementById("question");
const answerChoicesContainer = document.getElementById("answerChoices");

function displayQuestion(questionData) {
  document.getElementById("preGameLobby").classList.add("hidden");
  document.getElementById("gameContent").classList.remove("hidden");
  startGameTimer();

  // Display the question
  question.innerText = questionData.Question;

  // Clear previous answer choices
  answerChoicesContainer.innerHTML = "";

  if (!gameAdmin) {
    // Create answer choice elements and apply styling
    let colors = ["red", "blue", "yellow", "green"];
    questionData.Answers.forEach((answer, index) => {
      const answerChoiceDiv = document.createElement("div");
      answerChoiceDiv.className = `p-4 w-32 h-[200px] text-center rounded-md cursor-pointer bg-${colors[index]}-500 text-white`;

      // Add a click event listener to send the selected answer to the backend
      answerChoiceDiv.addEventListener("click", () => {
        sendAnswer(answer);
        // setTimeout(() => {
        //     ws.send(JSON.stringify(createClientMessage("score", "?")))
        //     console.log("asked how many points")
        // }, 1000);
      });

      answerChoiceDiv.textContent = answer;
      answerChoicesContainer.appendChild(answerChoiceDiv);
    });
  }

  // Show the "gameContent" section
  gameContent.classList.remove("hidden");
}

function sendAnswer(selectedAnswer) {
  answerChoicesContainer.innerHTML = "";
  answerChoicesContainer.innerText = "Answer sent to the server for grading";
  answerChoicesContainer.classList.add(["animate-ping", "animate-infinite", "animate-ease-in"]);
  question.innerText = "Waiting";
  ws.send(
    JSON.stringify(
      createClientMessage(
        "answer",
        JSON.stringify({
          answer: selectedAnswer,
          timerWhenAnswered: timerDisplay.innerText,
        })
      )
    )
  );
}
function populateScoreboard(data, actionButtonNotAllowed) {
  gameContent.classList.add("hidden");
  const scoreboardList = scoreboard.querySelector("ul");

  // Clear the existing entries
  scoreboardList.innerHTML = "";

  // Loop through the data and create entries
  data.forEach((entry, index) => {
    if (index < 5) {
      // Display only the top 5 entries
      const listItem = document.createElement("li");
      listItem.classList.add("border", "p-2", "flex", "items-center", "space-x-4");

      // Add profile picture
      const profilePicture = document.createElement("img");
      profilePicture.src = "data:image/jpeg;base64," + entry["picture"];
      profilePicture.alt = "Profile Picture";
      profilePicture.classList.add("w-8", "h-8", "rounded-full");
      listItem.appendChild(profilePicture);

      // Add player name and score
      const playerName = document.createElement("span");
      playerName.textContent = entry["name"];
      const playerScore = document.createElement("span");
      playerScore.textContent = `Score: ${entry["points"]}`;
      listItem.appendChild(playerName);
      listItem.appendChild(playerScore);

      scoreboardList.appendChild(listItem);
    }
  });
  scoreboard.classList.remove("hidden");
  if (gameAdmin) {
    actionButton.innerText = "continue";
    actionButton.classList = "block bg-blue-400 p-5 rounded-xl absolute top-20 right-2 animate-jump-in animate-ease-in";
  }
  if (actionButtonNotAllowed) {
    actionButton.classList = "hidden bg-blue-400 p-5 rounded-xl absolute top-20 right-2 animate-jump-in animate-ease-in";
  }
}
