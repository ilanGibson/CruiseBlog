const form = document.getElementById("blogForm");
const input = document.getElementById("blogInput")

form.addEventListener("submit", function (event) {
  event.preventDefault();
  console.log(input.value);
  input.value = "";
});
