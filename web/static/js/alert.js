function Alert(type, message, title = "") {
  if (type == null || type === "") {
    throw new Error("alert type may not be empty")
  } else if (message == null || message === "") {
    throw new Error("alert message may not be empty")
  }

  let template = document.getElementById('alert-template').content;
  let fragment = document.importNode(template,true);
  let alert = fragment.querySelector(".alert");
  alert.classList.add("alert-" + type);
  let heading = alert.querySelector(".alert-heading");
  if (title !== "") {
    heading.innerText = title;
  } else {
    switch (type) {
    case "danger":
      heading.innerText = "An error has occurred!"
      break;
    case "warning":
      heading.innerText = "Warning!"
        break;
    default:
      heading.classList.add("d-none");
    }
  }
  let msg = alert.querySelector("p");
  msg.innerText = message;

  let alerts = document.getElementById("alerts");
  alerts.appendChild(alert);
}