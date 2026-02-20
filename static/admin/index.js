const totalUsers = document.getElementById("totalUsers");
const totalPosts = document.getElementById("totalPosts");

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

connect()
