document.addEventListener('DOMContentLoaded', () => {
  const storeSecretBtn = document.getElementById('storeSecretBtn');
  const retrieveSecretBtn = document.getElementById('retrieveSecretBtn');

  // Store a new secret
  storeSecretBtn.addEventListener('click', () => {
      const secretText = document.getElementById('secret').value;
      if (secretText.trim() === "") {
          alert("Please enter a secret!");
          return;
      }

      // Send the secret to the backend
      fetch('/store', {
          method: 'POST',
          headers: {
              'Content-Type': 'text/plain'
          },
          body: secretText
      })
      .then(response => response.text())
      .then(data => {
          // Show the secret URL
          const secretResult = document.getElementById('secret-result');
          const secretLink = document.getElementById('secretLink');
          secretLink.textContent = `${data.trim()}`;
          secretLink.href = `${data.trim()}`;
          secretResult.style.display = 'block';
      })
      .catch(error => console.error('Error storing secret:', error));
  });

  // Retrieve a secret
  retrieveSecretBtn.addEventListener('click', () => {
      const secretKey = document.getElementById('secretKey').value.trim();
      if (secretKey === "") {
          alert("Please enter a secret key or URL!");
          return;
      }

      // If the input is a URL, extract the key
      const key = secretKey.includes("/secret/") ? secretKey.split("/secret/")[1] : secretKey;

      // Fetch the secret from the backend
      fetch(`/secret/${key}`)
      .then(response => {
          if (response.ok) {
              return response.text();
          } else {
              throw new Error('Secret not found or expired');
          }
      })
      .then(secret => {
          const retrievedSecret = document.getElementById('retrieved-secret');
          const secretText = document.getElementById('secretText');
          secretText.textContent = secret;
          retrievedSecret.style.display = 'block';
      })
      .catch(error => {
          alert(error.message);
      });
  });
});
