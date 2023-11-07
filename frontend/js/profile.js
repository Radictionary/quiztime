var nameInput = document.getElementById("name-input");
var emailInput = document.getElementById("email-input");
var nameSpan = document.getElementById("name-span");
var emailSpan = document.getElementById("email-span");
const applyButton = document.getElementById("applyButton");
if (emailInput != null) {
  emailInput.style.display == "none";
}
nameInput.style.display == "none";
let CSRFToken = "{{.CSRFToken}}";
applyButton.style.display = "none";
function toggleEdit(element) {
  var span = element.parentNode.querySelector("span");
  var input = element.parentNode.querySelector("input");
  if (span.style.display === "none") {
    span.style.display = "inline";
    input.style.display = "none";
    applyButton.style.display = "none";
  } else {
    span.style.display = "none";
    input.style.display = "inline";
    input.style.width = element.style.width
    applyButton.style.display = "block";
  }
  if ((emailInput != null && emailInput.style.display == "inline") || nameInput.style.display == "inline") {
    applyButton.style.display = "block";
  }
}

function applyChanges() {
  const userData = JSON.parse(accountJSON);

  userData.name = nameInput.value;
  if (userData.name.length === 0) {
    userData.name = "{{index .Account.Name}}";
  }
  if (emailInput != null) {
    userData.password = emailInput.value;
  }
  if (userData.password.length === 0) {
    userData.password = "{{index .Account.Password}}";
  }

  nameSpan.textContent = userData.name;
  if (emailSpan != null) {
    emailSpan.textContent = userData.password;
  }
  nameInput.placeholder = "{{index .Account.Name}}";

  if (userData.name != "{{index .Account.Name}}") {
    showNotificationWithConfirmation(
      `This will change your login information. The next time you want to login, your name will be ${userData.name} instead of ${name}`,
      function () {
        fetch(`/accounts/${name}`, {
          method: "PUT",
          body: JSON.stringify(userData),
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": CSRFToken,
          },
        })
          .then(function (response) {
            if (response.ok) {
              showNotification("Changes saved successfully");
              applyButton.style.display = "none";
              let buttons = document.querySelectorAll("button");
              for (var i = 0; i < buttons.length; i++) {
                var span = buttons[i].parentNode.querySelector("span");
                var input = buttons[i].parentNode.querySelector("input");
                if (span && input) {
                  span.style.display = "inline";
                  input.style.display = "none";
                }
              }
            } else {
              showNotification("Failed to update your settings");
            }
          })
          .catch(function (error) {
            showNotification("Changes saved, however errors occured");
            console.log(error);
          });
      },
      function () {
        emailSpan.textContent = "{{index .Account.Password}}";
        emailSpan.placeholder = "{{index .Account.Password}}";
        emailInput.textContent = "{{index .Account.Password}}";
        emailInput.value = "";
        nameSpan.textContent = "{{index .Account.Name}}";
        nameSpan.placeholder = "{{index .Account.Name}}";
        nameInput.value = "";
        applyButton.style.display = "none";
        let buttons = document.querySelectorAll("button");
        for (var i = 0; i < buttons.length; i++) {
          var span = buttons[i].parentNode.querySelector("span");
          var input = buttons[i].parentNode.querySelector("input");
          if (span && input) {
            span.style.display = "inline";
            input.style.display = "none";
          }
        }
      }
    );
  } else {
    fetch(`/accounts/${name}`, {
      method: "PUT",
      body: JSON.stringify(userData),
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": CSRFToken,
      },
    })
      .then(function (response) {
        if (response.ok) {
          showNotification("Changes saved successfully");
          applyButton.style.display = "none";
          let buttons = document.querySelectorAll("button");
          for (var i = 0; i < buttons.length; i++) {
            var span = buttons[i].parentNode.querySelector("span");
            var input = buttons[i].parentNode.querySelector("input");

            if (span && input) {
              span.style.display = "inline";
              input.style.display = "none";
            }
          }
        } else {
          showNotification("Failed to update your settings");
        }
      })
      .catch(function (error) {
        alert("An error occurred while applying changes.");
        console.log(error);
      });
  }
}
