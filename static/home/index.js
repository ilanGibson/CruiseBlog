const bannerAlert = document.getElementById("bannerAlert");
const homeDateLabel = document.getElementById("date");
const today = new Date();
const form = document.getElementById("blogForm");
const input = document.getElementById("blogInput");
const blogContainer = document.getElementById("blog-container");

if (window.location.pathname === "/home/") {
  "hitting 1"
  listPosts();
}

homeDateLabel.textContent = today.toDateString();

form.addEventListener("submit", function (event) {
  event.preventDefault();

  // input length check
  if (input.value.length < 1) {
    bannerAlert.textContent = "Input cannot be empty"
    bannerAlert.style.display='block';
    return
  } else if (input.value.length > 0 && bannerAlert.style.display=='block') {
    bannerAlert.textContent = ""
    bannerAlert.style.display='none'
  }

  fetch('/api/posts', {
    method: 'POST',
    headers: {
    'Content-Type': 'application/json'
    },
    body: JSON.stringify({content : input.value})
  })
  .then(function(res) {
      if (res.status === 422) {
        input.value = "";
        bannerAlert.textContent = "Input has been moderated";
        bannerAlert.style.display = 'block';
        throw new Error
      }
      return res.json()
  })
  .then(data => {
    const today = new Date();
    newPost(data.date, data.username, data.content);

    input.value = "";
  })
  .catch(function(err) {
    console.error("form submit err",err)
  })
});

function listPosts() {
  fetch('/api/posts')
    .then(res => {
      if (!res.ok) {
        // throw new Error('network res was not ok ' + response.status);
        window.location.replace('/');
      }
      return res.json();
    })
    .then(data => {
      if (data != null) {
        blogContainer.replaceChildren();
        for (let i = 0; i < data.length; i++) {
          newPost(data[i].date, data[i].username, data[i].content);
        }
      }
    })
    .catch(err => {
      console.error("list posts err", err)
    })
}

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
