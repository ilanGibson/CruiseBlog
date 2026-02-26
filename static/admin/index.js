const totalUsers = document.getElementById("totalUsers");
const totalPosts = document.getElementById("totalPosts");
const contentContainer = document.getElementById("content");
const postContainer = document.getElementById("postContainer");
const editPostBtn = document.getElementById("editPost-btn");
const deletePostBtn = document.getElementById("deletePost-btn");

let activeSection;
let activeSectionUsername;
let activeSectionTimeDate;
let newContentForUpdate;

function listPosts() {
  fetch('/api/posts')
    .then(res => {
      return res.json();
    })
    .then(data => {
      if (!data) return
      postContainer.replaceChildren();
      for (let i = 0; i < data.length; i++) {
        newPost(data[i].date, data[i].username, data[i].content);
      }
    })
    .catch(err => {
      console.error("list posts admin err", err)
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

  postContainer.prepend(newPost);
}


function connect() {
  const sseTotalUsers = new EventSource('/admin/sseEvents');

  sseTotalUsers.onopen = () => {
    console.log('connected to sse');
  };

  sseTotalUsers.onmessage = (event) => {
    const sseData = JSON.parse(event.data);
    totalUsers.textContent = sseData.total_users;
    totalPosts.textContent = sseData.total_posts;
  };

  sseTotalUsers.onerror = (err) => {
    console.error('sse error:', err);
  };
}

if (window.location.pathname === "/admin/") {
  connect();
  listPosts();
}

postContainer.addEventListener("click", (e) => {
  activeSection = e.target.closest("section");
  if (!activeSection) return;

  postContainer.querySelectorAll("section")
    .forEach(s => {
      s.classList.remove("active")
      });

  activeSection.classList.add("active");
  activeSectionTimeDate = activeSection.children[0].textContent;
  activeSectionUsername = activeSection.children[1].textContent.slice(0,8);
  contentContainer.value = activeSection.children[1].textContent.slice(10);
})

editPostBtn.addEventListener("click", (e) => {
  if (!activeSection) return;
  e.preventDefault();
  newContentForUpdate = contentContainer.value;

  fetch('/admin/post/update', {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      date: activeSectionTimeDate,
      username: activeSectionUsername,
      content: newContentForUpdate
    })
  })
  .then(res => {
      if (!res.ok) {
        throw new Error(`http error! status ${res.status}`);
      }
      const writeContent = `${activeSectionUsername}: ${contentContainer.value}`
      activeSection.children[1].textContent = writeContent;
      activeSection.classList.remove("active");
      contentContainer.value = "";
    })
  .catch(error => {
      console.error("error updating post:", error);
    });
});

deletePostBtn.addEventListener("click", (e) => {
  if (!activeSection) return;
  e.preventDefault();

  fetch('/admin/post/delete', {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      date: activeSectionTimeDate,
      username: activeSectionUsername,
    })
  })
  .then(res => {
    if (!res.ok) {
      throw new Error(`http error! status ${res.status}`);
    }
    activeSection.remove();
    contentContainer.value = ""
  })
  .catch(error => {
      console.error("error deleting post:", error);
    })
})
