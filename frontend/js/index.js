document.addEventListener("DOMContentLoaded", function () {
  const hamburgerButton = document.querySelector("[data-te-collapse-init]");
  const navbarContent = document.querySelector("[data-te-collapse-item]");

  let closed = false;

  hamburgerButton.addEventListener("click", function () {
    if (closed) {
      //closed
      document.getElementById("userLogo").style.width = "25px";
      document.getElementById("userLogo").style.height = "25px";
    } else {
      //open
      document.getElementById("userLogo").style.width = "50px";
      document.getElementById("userLogo").style.height = "50px";
    }
    closed = !closed;
  });
});

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

  var yesButton = document.createElement("button");
  yesButton.textContent = "Yes";
  yesButton.className = "btn btn-danger";
  yesButton.addEventListener("click", function () {
    onYes();
    notification.remove();
    overlay.remove();
  });
  buttonsContainer.appendChild(yesButton);

  var noButton = document.createElement("button");
  noButton.textContent = "No";
  noButton.className = "btn btn-info";
  noButton.addEventListener("click", function () {
    onNo();
    notification.remove();
    overlay.remove();
  });
  buttonsContainer.appendChild(noButton);

  notification.appendChild(buttonsContainer);

  document.body.appendChild(notification);
}
