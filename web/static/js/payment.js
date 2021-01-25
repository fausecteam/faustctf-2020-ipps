
window.addEventListener("load", function(){
  let form = document.getElementById("add-credit-card-form");
  let btn = form.querySelector("button[type=submit]");

  btn.addEventListener("click", addCreditCard);
});