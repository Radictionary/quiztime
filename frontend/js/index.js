// document.addEventListener("DOMContentLoaded", function () {
//   const hamburgerButton = document.querySelector("[data-te-collapse-init]");
//   const navbarContent = document.querySelector("[data-te-collapse-item]");

//   hamburgerButton.addEventListener("click", function () {
//     // Toggle the visibility of the navbar content
//     navbarContent.classList.toggle("hidden");
//     navbarContent.classList.toggle("flex");
//   });
// });

function showNotification(message, time) {
  if (time == undefined || time == null || time == 0) {
    time = 4000
  }
  var notification = document.createElement("div");
  notification.classList.add("notification");
  notification.textContent = message;

  document.body.appendChild(notification);

  setTimeout(function () {
    notification.remove();
  }, time); 
  return notification;
}
function showNotificationWithConfirmation(message, onYes, onNo) {
  var overlay = document.createElement("div");
  overlay.classList.add("overlay");
  document.body.appendChild(overlay);

  var notification = document.createElement("div");
  notification.classList.add("notification");
  notification.textContent = message;

  var buttonsContainer = document.createElement("div");
  buttonsContainer.classList.add("notification-buttons");
  buttonsContainer.classList.add("flex");
  buttonsContainer.classList.add("justify-between");

  var yesButton = document.createElement("button");
  yesButton.textContent = "Yes";
  yesButton.className = "btn mx-10 bg-slate-50 p-2 rounded-md";
  yesButton.addEventListener("click", function () {
    onYes();
    notification.remove();
    overlay.remove();
  });
  buttonsContainer.appendChild(yesButton);

  var noButton = document.createElement("button");
  noButton.textContent = "No";
  noButton.className = "btn btn-danger mx-10 bg-slate-50 p-2 rounded-md";
  noButton.addEventListener("click", function () {
    onNo();
    notification.remove();
    overlay.remove();
  });
  buttonsContainer.appendChild(noButton);

  notification.appendChild(buttonsContainer);

  document.body.appendChild(notification);
}
