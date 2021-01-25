window.addEventListener("load", function() {
  let form = document.getElementById("add-address-form");
  let btn = form.querySelector("button[type=submit]");

  btn.addEventListener("click", addAddress);
});
