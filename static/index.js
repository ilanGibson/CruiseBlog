const form = document.getElementById("blogForm");
const input = document.getElementById("blogInput")

form.addEventListener("submit", function (event) {
  event.preventDefault();

  fetch('http://localhost:8090/api/posts', {
  method: 'POST',
  headers: {
  'Content-Type': 'applicatioin/json'
  },
  body: JSON.stringify({content : input.value})
  })
  .then(input.value = "")
  .then(res => res.json())
  .then(data => {
      console.log(data);
      input.value = "";
    })
  .catch(err => {
      console.error("hi",err)
    })
});
