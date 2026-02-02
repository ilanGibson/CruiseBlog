const form = document.getElementById("blogForm");
const input = document.getElementById("blogInput");
const blogContainer = document.getElementById("blog-container");

form.addEventListener("submit", function (event) {
  event.preventDefault();

  fetch('http://localhost:8090/api/posts', {
  method: 'POST',
  headers: {
  'Content-Type': 'application/json'
  },
  body: JSON.stringify({content : input.value})
  })
  .then(res => res.json())
  .then(data => {
      const today = new Date();
      newPost(today, data.username, data.content)


      input.value = ""
    })
  .catch(err => {
      console.error("hi",err)
    })
});

function newPost(dateOfPost, postUsername, postContent) {
  const newPost = document.createElement('section');
  newPost.classList.add('blog-post');

  const date = document.createElement('p');
  date.classList.add('date');
  date.textContent = dateOfPost;

  const content = document.createElement('p');
  content.classList.add('content');
  content.textContent = `${postUsername}: ${postContent}`

  newPost.appendChild(date);
  newPost.appendChild(content);

  // add new post to the top
  blogContainer.prepend(newPost);
}
