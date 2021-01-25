window.addEventListener("load", function() {
  let form = document.getElementById("login-form");
  let btn = form.querySelector("button[type=submit]");

  btn.addEventListener("click", login);
});