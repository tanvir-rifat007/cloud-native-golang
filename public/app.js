import { API } from "./services/API.js";
import Router from "./services/Router.js";
import proxiedStore from "./services/Store.js";

globalThis.addEventListener("DOMContentLoaded", () => {
  app.router.init();
});

globalThis.app = {
  api: API,

  signupNewsletter: async (event) => {
    const email = document.querySelector("#email").value;

    event.preventDefault();

    const res = await API.postNewsletter({
      email,
    });

    console.log(res);

    const modal = document.getElementById("modal");

    modal.showModal();

    // removing the previous error message
    const errorMessage = modal.querySelector(".error-message");
    if (errorMessage) {
      errorMessage.remove();
    }

    if (!res.error) {
      app.store.token = res.token;

      const modalContent = document.getElementById("modal-content");
      const modalTitle = document.getElementById("modal-title");

      modalContent.innerHTML = `<p class="text-gray-300 text-center">You have successfully signed up for the newsletter!</p>`;
      modalTitle.innerHTML = `<h3 class="text-xl font-semibold mb-2 text-white">Success</h3>`;
      modalContent.className = "text-gray-300";
    }
    if (res.error) {
      const modalContent = document.getElementById("modal-content");
      const modalTitle = document.getElementById("modal-title");

      modalContent.innerHTML = `<p class="text-gray-300 text-center">There was an error signing up for the newsletter. Please try again later.</p>`;
      modalTitle.innerHTML = `<h3 class="text-xl font-semibold mb-2 text-white">Error</h3>`;
      modalContent.className = "text-gray-300";

      const message = res.error.email || res.error;
      const p = document.createElement("p");
      p.innerText = message;
      // tailwind classes
      p.className = "text-red-500 text-center mt-2 error-message";

      modal.appendChild(p);
    }

    // Array.from(modal.querySelectorAll(".meal-modal, .workout-modal")).forEach(
    //   (el) => el.remove()
    // );

    // if (!res.meals) {
    //   const mealElements = document.createElement("div");
    //   mealElements.classList.add("meal-modal");
    //   mealElements.innerHTML = `
    //     <p>No workouts found</p>
    //   `;
    //   modal.appendChild(mealElements);
    //   return;
    // }

    // res.meals.forEach((meal) => {
    //   const mealElements = document.createElement("div");
    //   mealElements.classList.add("meal-modal");
    //   mealElements.innerHTML = `

    //     <p>Name : ${meal.name} <span style="color:hsl(20, 90%, 60%)">(${meal.calories}cal)</span></p>
    //     <p>Description: ${meal.description}</p>
    //     <br/>
    //   `;
    //   modal.appendChild(mealElements);
    // });

    // console.log(res);
  },

  searchNewsletters: async (event) => {
    const query = document.querySelector("#search").value;

    event.preventDefault();

    const response = await API.searchNewsletters({
      query,
    });

    console.log(response);

    app.router.go(`/newsletters/search?query=${query}`);
  },

  showError: (response = "There was an error", goToHome = false) => {
    const modal = document.getElementById("modal");

    const modalContent = document.getElementById("modal-content");
    const modalTitle = document.getElementById("modal-title");
    const modalBtn = document.querySelector("#modal button");

    modalContent.innerHTML = `<p class="text-gray-300 text-center">There was an error signing up for the newsletter. Please try again later.</p>`;
    modalTitle.innerHTML = `<h3 class="text-xl font-semibold mb-2 text-white">Error</h3>`;
    modalContent.className = "text-gray-300";

    modal.showModal();

    const message = response?.error?.email || response?.error || response;
    const p = document.createElement("p");
    p.innerText = message;
    p.style.fontSize = "1.5rem";
    p.classList.add("error-message");

    modal.appendChild(p);

    if (goToHome) {
      setTimeout(() => {
        app.router.go("/");
        app.closeModal();
      }, 3000);
    }
  },

  router: Router,

  closeModal: () => {
    document.querySelector("#modal").close();
  },

  store: proxiedStore,
};
