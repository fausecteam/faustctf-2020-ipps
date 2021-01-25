function login(event) {
  event.preventDefault();

  let form = document.getElementById("login-form");
  let data = new FormData(form);

  event.target.disabled = true;
  let spinner = event.target.querySelector(".spinner-border");
  spinner.classList.remove("d-none");

  fetch("/api/login", {
    method: "POST",
    body: data,
  }).then((raw) => raw.json()).then((json) => {
    if (json.error != null && json.error !== "") {
      throw new Error(json.error);
    }

    sessionStorage.setItem('username', json.result.toString())
    window.location.replace("/");
  }).catch((reason) => {
    new Alert("danger", reason.message);
  }).finally(() => {
    spinner.classList.add("d-none");
    event.target.disabled = false;
  });
}

function getUsername() {
  let username = sessionStorage.getItem('username');
  if (username != null && username !== "") {
    return Promise.resolve(username);
  }

  return fetch("/api/login", {
    method: "POST",
    body: new FormData(),
  }).then((raw) => raw.json()).then((response) => {
    if (response.error != null && response.error !== "") {
      throw new Error(
          "Your session has expired. Please log in again to resolve this issue.");
    }

    sessionStorage.setItem('username', response.result);
    return response.result;
  });
}

function addAddress(event) {
  event.preventDefault()

  let form = document.getElementById("add-address-form");
  let data = new FormData(form);

  getUsername().then((username) => {
    event.target.disabled = true;
    let spinner = event.target.querySelector(".spinner-border");
    spinner.classList.remove("d-none");

    fetch("/api/user/" + username + "/add-address", {
      method: "POST",
      body: data,
    }).then((raw) => raw.json()).then((response) => {
      if (response.error != null && response.error !== "") {
        new Alert("danger", response.error);
        return;
      }

      new Alert("success", "Your address has been added successfully!");
      reloadAddresses();
    }).catch((reason) => {
      new Alert("danger", reason.message);
    }).finally(() => {
      spinner.classList.add("d-none");
      event.target.disabled = false;
    });
  });
}

function reloadAddresses() {
  getUsername().then((username) => {
    fetch("/api/user/" + username + "/get-addresses")
    .then((raw) => raw.json()).then((response) => {
      if (response.error != null && response.error !== "") {
        new Alert("danger", response.error);
        return;
      }
      let addresses = document.querySelector("#addresses tbody");
      addresses.innerHTML = "";
      for (let address of response.result) {
        let row = document.createElement("tr");
        row.innerHTML = `<td>${address.street}</td>
        <td>${address.zip}</td>
        <td>${address.city}</td>
        <td>${address.country}</td>
        <td>${address.planet}</td>`;

        addresses.appendChild(row);
      }
    });
  });

}

function addCreditCard(event) {
  event.preventDefault();

  let form = document.getElementById("add-credit-card-form");
  let data = new FormData(form);

  getUsername().then((username) => {
    event.target.disabled = true;
    let spinner = event.target.querySelector(".spinner-border");
    spinner.classList.remove("d-none");

    fetch("/api/user/"+ username + "/add-credit-card", {
      method: "POST",
      body: data,
    }).then((raw) => raw.json()).then((response) => {
      if (response.error != null && response.error !== "") {
        new Alert("danger", response.error);
        return;
      }

      new Alert("success", "Your credit card has been added successfully!");
      reloadCreditCards();
    }).catch((reason) => {
      new Alert("danger", reason.message);
    }).finally(() => {
      spinner.classList.add("d-none");
      event.target.disabled = false;
    });
  });
}

function reloadCreditCards() {
  getUsername().then((username) => {
    fetch("/api/user/" + username + "/get-credit-cards")
    .then((raw) => raw.json()).then((response) => {
      if (response.error != null && response.error !== "") {
        new Alert("danger", response.error);
        return;
      }
      let creditCards = document.querySelector("#credit-cards tbody");
      creditCards.innerHTML = "";
      for (let card of response.result) {
        let row = document.createElement("tr");
        row.innerHTML = `
          <td>MarsCard</td>
          <td>${card.number}</td>
          <td>TODO</td>`;
        creditCards.appendChild(row);
      }
    });
  });
}
