export class SearchNewslettersPage extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
  }

  async render() {
    const homePageTemplate = document.getElementById(
      "search-newsletter-page-template"
    );
    console.log(homePageTemplate);
    const templateContent = homePageTemplate.content.cloneNode(true);
    this.appendChild(templateContent);

    const query = new URLSearchParams(window.location.search);
    console.log(query);
    const newsletters = (await app.api.searchNewsletters(query)).newsletters;
    console.log("newsletters,", newsletters);

    if (newsletters.length === 0) {
      const newslettersContainer = this.querySelector("#newsletters-container");

      const newsletterItem = document.createElement("div");

      newsletterItem.innerHTML = `
  <h3 class="text-xl font-semibold mb-2 text-white">No newsletters found</h3>
      `;
      newslettersContainer.appendChild(newsletterItem);

      return;
    }

    const newslettersContainer = this.querySelector("#newsletters-container");

    newsletters.forEach((newsletter) => {
      const newsletterItem = document.createElement("div");
      // style this using tailwind css
      newsletterItem.className = `
  bg-white/10 backdrop-blur-md border border-white/20 
  rounded-2xl p-6 mb-6 
  shadow-xl transition-transform transform 
  hover:scale-105 hover:shadow-2xl
  sm:w-[50%]
  sm:max-w-[400px] mx-auto
  w-full
  flex flex-col
  mt-10



  
`;

      newsletterItem.innerHTML = `
  <h3 class="text-xl font-semibold mb-2 text-white">${newsletter.Title}</h3>
  <p class="text-sm text-gray-300 mb-4">
   created on : ${new Date(newsletter.CreatedAT).toLocaleDateString("en-US", {
     year: "numeric",
     month: "long",
     day: "numeric",
   })} at ${new Date(newsletter.CreatedAT).toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      })}
   
  </p>
  <a onclick= "app.router.go('/newsletters/${newsletter.ID}')"

     class="inline-block text-sm font-medium text-indigo-300 hover:text-indigo-400 transition-colors cursor-pointer">
    Read more →
  </a>
`;

      newslettersContainer.appendChild(newsletterItem);
    });
  }
}

customElements.define("search-newsletter-page", SearchNewslettersPage);
