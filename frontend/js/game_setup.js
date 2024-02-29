let gameData = JSON.parse(gameJSON);
console.log("GAME DATA IS:", gameData);
let gamePicture = ""
gamePicture = gameData.GamePicture;
let usersShared = gameData.UsersShared;
if ((gameData.Owner != null && gameData.DateOfCreation != "") || (gameData.Owner != "" && gameData.DateOfCreation != "")) {
  document.getElementById("owner").value = gameData.Owner;
}
if (gameData.Timer != 0) {
  document.getElementById("timer").value = gameData.Timer;
}

let currentUsersShared = usersShared;
// document.getElementById("gamePicture").src = `data:image/jpeg;base64,${gameData.GamePicture}`;
if (gameData.GamePicture != "" && gameData.GamePicture != null) {
  // document.querySelector(".profile-pic").src = `data:image/jpeg;base64,${gameData.GamePicture}`;
  document.querySelector(".bi-card-image").style.display = "none";
  document.querySelector(".profile-pic").style.display = "initial";
  document.querySelector(".profile-pic").src = `data:image/jpeg;base64,${gameData.GamePicture}`;
  document.body.style.backgroundImage = `url(data:image/jpeg;base64,${gameData.GamePicture})`;
  // document.body.style.backgroundRepeat = "no-repeat"
} else {
  document.getElementById("removePictureButton").style.display = "none";
}
let notification;
document.getElementById("navbar").addEventListener("mouseenter", () => {
  try {
    notification.remove();
  } catch (error) {}
  showNotification("Don't forget to save!", 1200);
});

function saveGame() {
  // Collect game details
  let gameDetails = {
    owner: document.getElementById("owner").value,
    gameName: document.getElementById("gameName").value,
    timer: parseInt(document.getElementById("timer").value),
  };

  if (parseInt(gameDetails.timer) < 5) {
    gameDetails.timer = 5;
    showNotification("Timer must be at least 5 seconds. Setting timer to 5 seconds");
  }

  // Collect question details
  const questions = [];
  const questionTabs = document.querySelectorAll(".tab-content");

  // Start the loop from index 1 to skip the first tab (game details tab)
  for (let index = 1; index < questionTabs.length; index++) {
    const questionTab = questionTabs[index];
    const questionData = {
      questionNumber: index - 1, // Subtract 1 to get the correct question number
      question: questionTab.querySelector('input[name="question"]').value,
      points: parseInt(questionTab.querySelector('input[name="points"]').value),
      // Extract correct answers based on checkboxes
      correctAnswers: [],
      answers: [questionTab.querySelector('input[name="correctAnswer"]').value, questionTab.querySelector('input[name="wrongAnswer1"]').value, questionTab.querySelector('input[name="wrongAnswer2"]').value, questionTab.querySelector('input[name="wrongAnswer3"]').value],
    };

    // Check checkboxes for correct answers and add them to the correctAnswers array
    if (questionTab.querySelector('input[name="isCorrect"]').checked) {
      questionData.correctAnswers.push(questionTab.querySelector('input[name="correctAnswer"]').value);
    }
    if (questionTab.querySelector('input[name="isCorrect1"]').checked) {
      questionData.correctAnswers.push(questionTab.querySelector('input[name="wrongAnswer1"]').value);
    }
    if (questionTab.querySelector('input[name="isCorrect2"]').checked) {
      questionData.correctAnswers.push(questionTab.querySelector('input[name="wrongAnswer2"]').value);
    }
    if (questionTab.querySelector('input[name="isCorrect3"]').checked) {
      questionData.correctAnswers.push(questionTab.querySelector('input[name="wrongAnswer3"]').value);
    }

    questions.push(questionData);
  }

  let game;
  // Create the game object
  if (gameData.DateOfCreation == "") {
    game = {
      name: gameDetails.gameName,
      owner: document.getElementById("owner").value,
      status: "active",
      timer: gameDetails.timer,
      questionsAnswers: questions,
      players: {},
      questions: questions.length,
      gamePicture: gamePicture,
    };
  } else {
    game = {
      name: gameDetails.gameName,
      owner: gameDetails.owner,
      usersShared: currentUsersShared,
      status: "active",
      timer: gameDetails.timer,
      questionsAnswers: questions,
      players: {},
      questions: questions.length,
      gamePicture: gamePicture,
    };
  }

  fetch("/game/" + game.name, {
    method: "PUT", 
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(game),
  })
    .then((response) => {
      if (response.ok) {
        if (document.querySelector(".profile-pic").src == "" || document.querySelector(".bi-card-image").style.display == "initial" || gamePicture == "") {
          document.getElementById("removePictureButton").style.display = "none";
        } else {
          document.getElementById("removePictureButton").style.display = "block";
        }
        showNotification("Saved Game");
      } else {
        alert("Failed to save the game:" + response.status);
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}
function openShareModal() {
  const modalBackdrop = document.getElementById("modal-backdrop");
  const shareModal = document.getElementById("share-modal");

  modalBackdrop.classList.remove("hidden");
  shareModal.classList.remove("hidden");
}
function closeShareModal() {
  const modalBackdrop = document.getElementById("modal-backdrop");
  const shareModal = document.getElementById("share-modal");

  modalBackdrop.classList.add("hidden");
  shareModal.classList.add("hidden");
}

// Function to create tabs and content based on game data
function createTabsAndContent(data) {
  const tabsContainer = document.getElementById("tabs");
  const mainContent = document.querySelector("main");

  // if (data.DateOfCreation == "") {
  //   saveGame();
  //   return;
  // }
  data.QuestionsAnswers.forEach((question, index) => {
    const tabId = `question-${index}`;
    const tabListItem = document.createElement("li");
    tabListItem.textContent = `Question ${index + 1}`;
    tabListItem.className = "cursor-pointer";
    tabListItem.setAttribute("data-tab", tabId);
    tabListItem.onclick = () => showTab(tabId);
    tabsContainer.appendChild(tabListItem);

    const questionTab = document.createElement("section");
    questionTab.id = tabId;
    questionTab.className = "tab-content hidden";
    questionTab.innerHTML = `
            <h2 class="text-xl font-semibold mb-2">Question ${index + 1}</h2>
            <label for="question">Question:</label>
            <input type="text" id="question" name="question" class="w-[50%] border p-1 mb-2" placeholder="Enter question" value="${question.Question}">
            <label for="question">Points:</label>
            <input type="number" id="points" name="points" class="border p-1 mb-2" placeholder="Points" value="${question.Points}">
            <button class="bg-red-900 text-white p-2 rounded cursor-pointer mx-4" onclick="deleteQuestionTab('${tabId}', ${index})">Delete Question</button>

            <br>
            <div class="grid space-x-4 space-y-4 grid-cols-2 grid-row-2">
                <!-- Box for Correct Answer -->
                <div class="bg-red-400 p-2 rounded-md flex-1 h-[200px] mt-4 ml-4">
                    <label for="correctAnswer">Answer 1:</label>
                    <input type="text" id="correctAnswer" name="correctAnswer" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 1" value="${question.Answers[0]}">
                    <input type="checkbox" id="isCorrect" name="isCorrect" value="true" ${question.CorrectAnswers.includes(question.Answers[0]) ? "checked" : ""}> Correct
                </div>

                <!-- Boxes for Wrong Answers -->
                <div class="bg-blue-500 p-4 rounded-md flex-1 h-[200px]">
                    <label for="wrongAnswer1">Answer 2:</label>
                    <input type="text" id="wrongAnswer1" name="wrongAnswer1" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 2" value="${question.Answers[1]}">
                    <input type="checkbox" id="isCorrect1" name="isCorrect1" value="true" ${question.CorrectAnswers.includes(question.Answers[1]) ? "checked" : ""}> Correct
                </div>

                <div class="bg-yellow-500 p-4 rounded-md flex-1 h-[200px]">
                    <label for="wrongAnswer2">Answer 3:</label>
                    <input type="text" id="wrongAnswer2" name="wrongAnswer2" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 3" value="${question.Answers[2]}">
                    <input type="checkbox" id="isCorrect2" name="isCorrect2" value="true" ${question.CorrectAnswers.includes(question.Answers[2]) ? "checked" : ""}> Correct
                </div>

                <div class="bg-green-500 p-4 rounded-md flex-1 h-[200px]">
                    <label for="wrongAnswer3">Answer 4:</label>
                    <input type="text" id="wrongAnswer3" name="wrongAnswer3" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 4" value="${question.Answers[3]}">
                    <input type="checkbox" id="isCorrect3" name="isCorrect3" value="true" ${question.CorrectAnswers.includes(question.Answers[3]) ? "checked" : ""}> Correct
                </div>
            </div>
        `;
    questionTab.style.backgroundColor = "#7c848af9";
    questionTab.style.padding = "1rem";
    mainContent.appendChild(questionTab);
  });

  // Show the first question tab by default
  showTab("game-details");
}
try {
  createTabsAndContent(gameData);
} catch (error) {
  saveGame()
  document.getElementById("removePictureButton").style.display = "none";
  showNotification("Initial game created and saved")
}

// JavaScript functions for handling tabs
function showTab(tabId) {
  const tabs = document.querySelectorAll(".tab-content");
  tabs.forEach((tab) => tab.classList.add("hidden"));
  if (document.getElementById(tabId) != null) {
    document.getElementById(tabId).classList.remove("hidden");
  } else {
    return;
  }

  const tabListItems = document.querySelectorAll("#tabs li");
  tabListItems.forEach((tabItem) => tabItem.classList.remove("selected"));

  if (tabId != "game-details") document.querySelector(`#tabs li[data-tab="${tabId}"]`).classList.add("selected");
}

function addNewQuestionTab() {
  const tabs = document.getElementById("tabs");
  const newQuestionNumber = tabs.children.length - 1; // Get the number of questions
  const newTab = document.createElement("li");
  const tabId = `question-${newQuestionNumber}`;
  newTab.textContent = `Question ${newQuestionNumber}`;
  newTab.className = "cursor-pointer";
  newTab.setAttribute("data-tab", tabId); // Set the data-tab attribute
  newTab.onclick = () => showTab(tabId);

  // Append the new tab to the end of the list
  tabs.appendChild(newTab);

  // Create a new question tab
  const questionTab = document.createElement("section");
  questionTab.id = tabId;
  questionTab.className = "tab-content hidden";
  questionTab.innerHTML = `
    <h2 class="text-xl font-semibold mb-2">Question ${newQuestionNumber}</h2>
    <label for="question-${newQuestionNumber}">Question:</label>
    <input type="text" id="question-${newQuestionNumber}" name="question" class="w-[50%] border p-1 mb-2" placeholder="Enter question">
    <label for="points-${newQuestionNumber}">Points:</label>
    <input type="number" id="points-${newQuestionNumber}" name="points" class="border p-1 mb-2" placeholder="Points" value="1000">

    <button class="bg-red-900 text-white p-2 rounded cursor-pointer mx-4" onclick="deleteQuestionTab('${tabId}', ${newQuestionNumber})">Delete Question</button>

    <br>
    <div class="grid space-x-4 space-y-4 grid-cols-2 grid-row-2">
        <div class="bg-red-500 p-2 rounded-md flex-1 h-[200px] mt-4 ml-4">
            <label for="correctAnswer-${newQuestionNumber}">Answer 1:</label>
            <input type="text" id="correctAnswer-${newQuestionNumber}" name="correctAnswer" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 1">
            <input type="checkbox" id="isCorrect-${newQuestionNumber}" name="isCorrect" value="true"> Correct
        </div>

        <div class="bg-blue-500 p-4 rounded-md flex-1 h-[200px]">129775
            <label for="wrongAnswer1-${newQuestionNumber}">Answer 2:</label>
            <input type="text" id="wrongAnswer1-${newQuestionNumber}" name="wrongAnswer1" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 2">
            <input type="checkbox" id="isCorrect1-${newQuestionNumber}" name="isCorrect1" value="true"> Correct
        </div>

        <div class="bg-yellow-500 p-4 rounded-md flex-1 h-[200px]">
            <label for="wrongAnswer2-${newQuestionNumber}">Answer 3:</label>
            <input type="text" id="wrongAnswer2-${newQuestionNumber}" name="wrongAnswer2" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 3">
            <input type="checkbox" id="isCorrect2-${newQuestionNumber}" name="isCorrect2" value="true"> Correct
        </div>

        <div class="bg-green-500 p-4 rounded-md flex-1 h-[200px]">
            <label for="wrongAnswer3-${newQuestionNumber}">Answer 4:</label>
            <input type="text" id="wrongAnswer3-${newQuestionNumber}" name="wrongAnswer3" class="block w-full border-0 p-2 bg-slate-200 rounded-lg mt-5 text-black" placeholder="Enter Answer 4">
            <input type="checkbox" id="isCorrect3-${newQuestionNumber}" name="isCorrect3" value="true"> Correct
        </div>
    </div>
`;
  questionTab.style.backgroundColor = "#7c848af9";
  questionTab.style.padding = "1rem";

  document.querySelector("main").appendChild(questionTab);

  // Show the newly created question tab
  showTab(tabId);
}

function deleteQuestionTab(tabId, questionNumber) {
  const tabToDelete = document.getElementById(tabId);
  if (tabToDelete) {
    tabToDelete.remove();
    // Remove the corresponding tab item in the sidebar
    const tabListItem = document.querySelector(`#tabs li[data-tab="${tabId}"]`);
    if (tabListItem) {
      tabListItem.remove();
    }
    // Select the previous question tab or the "Game Details" tab
    const prevTabId = `question-${questionNumber - 1}`;
    if (questionNumber - 1 == -1 || 0) {
      showTab("game-details");
    } else {
      showTab(prevTabId);
    }
  }
}

// Function to add a shared user to the shared box
function addSharedUserToBox(username) {
  const sharedUsersBox = document.getElementById("sharedUsersBox");
  const sharedUserItem = document.createElement("div");
  sharedUserItem.className = "flex items-center justify-between py-2 border-b";
  sharedUserItem.innerHTML = `
        <span>${username}</span>
        <button class="text-red-500 hover:text-red-700 cursor-pointer" onclick="removeSharedUser('${username}')">x</button>
    `;
  sharedUsersBox.appendChild(sharedUserItem);
}

if (usersShared != null) {
  usersShared.forEach((user) => {
    addSharedUserToBox(user);
  });
}

function shareGame() {
  const sharedUserInput = document.getElementById("sharedUser");
  const sharedUserName = sharedUserInput.value.trim();

  if (sharedUserName === "") {
    // Handle empty input
    const modalBackdrop = document.getElementById("modal-backdrop");
    const shareModal = document.getElementById("share-modal");

    modalBackdrop.classList.add("hidden");
    shareModal.classList.add("hidden");
    showNotification("Please enter a username to share the game.");
    return;
  }

  // Send a request to the backend to share the game with the entered user
  fetch(`/game/${gameName}/share?user=${sharedUserName}`, {
    method: "PUT",
  })
    .then((response) => {
      if (response.ok) {
        // Sharing successful, add the shared user to the list
        addSharedUser(sharedUserName);
        sharedUserInput.value = ""; // Clear the input field
      } else {
        // Handle errors, e.g., user not found
        alert("Failed to share the game. Please check the username.");
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function addSharedUser(username) {
  const sharedUsersBox = document.getElementById("sharedUsersBox");
  const sharedUserItem = document.createElement("div");
  sharedUserItem.className = "flex items-center justify-between py-2 border-b";
  sharedUserItem.innerHTML = `
        <span>${username}</span>
        <button class="text-red-500 hover:text-red-700 cursor-pointer" onclick="removeSharedUser('${username}')">x</button>
    `;
  sharedUsersBox.appendChild(sharedUserItem);
  if (currentUsersShared == null) {
    currentUsersShared = [];
  }
  currentUsersShared.push(username);
}

function removeSharedUser(username) {
  // Send a request to the backend to remove the shared user
  fetch(`/game/${gameName}/share?user=${username}`, {
    method: "DELETE",
  })
    .then((response) => {
      if (response.ok) {
        const sharedUsersBox = document.getElementById("sharedUsersBox");
        const sharedUserItems = sharedUsersBox.querySelectorAll("div");

        for (const sharedUserItem of sharedUserItems) {
          const usernameElement = sharedUserItem.querySelector("span");
          if (usernameElement.textContent === username) {
            sharedUserItem.remove();
            break; 
          }
        }
      } else {
        alert("Failed to remove the shared user.");
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function removeGamePicture() {
  gamePicture = "";
  saveGame();
  document.querySelector(".bi-card-image").style.display = "initial";
  document.querySelector(".profile-pic").style.display = "none";
  document.querySelector(".profile-pic").src = `#`;
  document.body.style.backgroundImage = "none";
}

function playGame() {
  saveGame();
  fetch(`/game/${gameName}/startgame?name=${userName}`, {
    method: "GET",
  })
    .then((response) => response.text())
    .then((text) => {
      if (text != "" && text != null) {
        location.replace(`http://${window.location.hostname}:${window.location.port}/play/${text}?name=${userName}`); //as the first person to join, will become admin
      }
    });
}
