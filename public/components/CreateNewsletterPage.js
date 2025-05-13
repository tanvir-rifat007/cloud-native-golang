import { API } from "../services/API.js";

export class CreateNewsletter extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
  }

  async render() {
    const homePageTemplate = document.getElementById(
      "create-newsletter-page-template"
    );
    console.log(homePageTemplate);
    const templateContent = homePageTemplate.content.cloneNode(true);
    this.appendChild(templateContent);

    const tagInput = document.getElementById("tag-input");
    const tagsList = document.getElementById("tags-list");
    const contentInput = document.getElementById("content");
    const preview = document.getElementById("markdown-preview");

    let tags = [];

    // --- Handle tag input ---
    tagInput.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && tagInput.value.trim() !== "") {
        e.preventDefault();
        const newTag = tagInput.value.trim();
        if (!tags.includes(newTag)) {
          tags.push(newTag);
          renderTags();
        }
        tagInput.value = "";
      }
    });

    function renderTags() {
      tagsList.innerHTML = "";
      tags.forEach((tag, index) => {
        const tagEl = document.createElement("span");
        tagEl.textContent = tag;
        tagEl.className = `
          bg-gray-200 text-gray-800 text-sm font-semibold mr-2 px-2.5 py-0.5 rounded
        `;

        const delBtn = document.createElement("button");
        delBtn.textContent = "×";
        delBtn.style =
          "margin-left:5px; background:none; border:none; cursor:pointer;";
        delBtn.onclick = () => {
          tags.splice(index, 1);
          renderTags();
        };

        tagEl.appendChild(delBtn);
        tagsList.appendChild(tagEl);
      });
    }

    // --- Markdown live preview ---
    contentInput.addEventListener("input", () => {
      const raw = contentInput.value;
      preview.innerHTML = marked.parse(raw);

      console.log(marked.parse(raw));
    });

    console.log(document.getElementById("submit-post"));
    document
      .getElementById("submit-post")
      .addEventListener("click", async (event) => {
        console.log(event);
        event.preventDefault();

        const title = document.getElementById("title").value;
        const markdown = contentInput.value;

        const payload = {
          title,
          body: markdown,
          tags,
          token: app.store.token ? app.store.token : null,
        };

        console.log(payload);

        const response = await API.fetchData("/newsletter/create", payload);

        console.log(response);

        // removing error message
        const errorMessages = document.querySelectorAll(".error-message");
        errorMessages.forEach((el) => el.remove());

        if (!response.error) {
          const modal = document.getElementById("modal");
          const modalContent = document.getElementById("modal-content");
          const modalTitle = document.getElementById("modal-title");

          modalContent.innerHTML = `<p class="text-gray-300 text-center">Newsletter created successfully.</p>`;
          modalTitle.innerHTML = `<h3 class="text-xl font-semibold mb-2 text-white">Success</h3>`;
          modalContent.className = "text-gray-300";

          modal.showModal();

          setTimeout(() => {
            modal.close();
            app.router.go("/newsletters");
          }, 2000);
        } else {
          const modal = document.getElementById("modal");
          const modalContent = document.getElementById("modal-content");
          const modalTitle = document.getElementById("modal-title");

          modalContent.innerHTML = `<p class="text-gray-300 text-center">There was an error creating the newsletter. Please try again later.</p>`;
          modalTitle.innerHTML = `<h3 class="text-xl font-semibold mb-2 text-white">Error</h3>`;
          modalContent.className = "text-gray-300";

          modal.showModal();

          let message;
          if (
            (response.error =
              "the server encountered a problem and could not process your request")
          ) {
            message = "You need to admin access to create a newsletter";
          } else {
            message = response.error.title || response.error;
          }
          const p = document.createElement("p");
          p.innerText = message;

          p.className = "text-red-500 text-center mt-2 error-message";

          modal.appendChild(p);
        }
      });
  }
}

customElements.define("create-newsletter-page", CreateNewsletter);
