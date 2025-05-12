export class HomePage extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
  }

  async render() {
    const homePageTemplate = document.getElementById("home-page-template");
    console.log(homePageTemplate);
    const templateContent = homePageTemplate.content.cloneNode(true);
    this.appendChild(templateContent);
  }
}

customElements.define("home-page", HomePage);
