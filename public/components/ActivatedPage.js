export class ActivatedPage extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
  }

  render() {
    const workoutPageTemplate = document.getElementById(
      "activated-page-template"
    );
    console.log(workoutPageTemplate);
    const templateContent = workoutPageTemplate.content.cloneNode(true);
    this.appendChild(templateContent);

    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get("token");

    if (token) {
      app.store.jwt = token;
    }
  }
}

customElements.define("activated-page", ActivatedPage);
