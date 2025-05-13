export class NewsletterPage extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
  }

  async render() {
    const homePageTemplate = document.getElementById(
      "newsletter-page-template"
    );
    console.log(homePageTemplate);
    const templateContent = homePageTemplate.content.cloneNode(true);
    this.appendChild(templateContent);

    // get the id from the url
    let id = window.location.pathname.split("/").pop();

    let newsletter = (await app.api.getNewsletter(id)).newsletter;

    console.log("newsletter", newsletter);

    // style this using tailwind css
    const newsletterContainer = this.querySelector("#newsletter-container");

    newsletterContainer.innerHTML = `
    <div class="container mx-auto mt-10">
      <img 
        src="https://canvas-assetsbucket-mzmt07gci4uk.s3.eu-north-1.amazonaws.com/${
          newsletter.FileURLs[0]
        }"
        alt="${newsletter.Title}"
        class="rounded-2xl mb-4 w-full h-auto object-cover"
      />
      <h3 class="text-xl font-semibold mb-2 text-white">${newsletter.Title}</h3>

      <p class="text-sm text-gray-300 mb-4">
        created By : <span
          class="text-white font-semibold"
        >${app.store.createdBy}</span> at ${new Date(
      newsletter.CreatedAT
    ).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    })} at ${new Date(newsletter.CreatedAT).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    })}
      </p>
      <p class="text-sm text-gray-300 mb-4">
        ${marked.parse(newsletter.Body)}
      </p>

      <div class="flex flex-wrap mt-4">

        
        
        ${newsletter?.Tags?.map((tag) => {
          // set the random background color

          const colors = [
            "bg-red-500",
            "bg-blue-500",
            "bg-green-500",
            "bg-yellow-500",
            "bg-purple-500",
            "bg-pink-500",
          ];

          const randomColor = colors[Math.floor(Math.random() * colors.length)];

          return `<span class="inline-flex items-center px-3 py-1 text-sm font-medium text-white rounded-full ${randomColor} mr-2 mb-2">${tag}</span>`;
        })}

    

          </div>

    `;
  }
}

customElements.define("newsletter-page", NewsletterPage);
