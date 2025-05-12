import { ActivatedPage } from "../components/ActivatedPage.js";
import { CreateNewsletter } from "../components/CreateNewsletterPage.js";
import { HomePage } from "../components/HomePage.js";
import { NewsletterPage } from "../components/NewsletterPage.js";
import { NewslettersPage } from "../components/NewslettersPage.js";
import { SearchNewslettersPage } from "../components/SearchNewsletterPage.js";

export const routes = [
  {
    path: "/",
    component: HomePage,
  },

  {
    path: "/activated",
    component: ActivatedPage,
  },

  {
    path: "/newsletters",
    component: NewslettersPage,
  },

  {
    path: /\/newsletters\/(\d+)/,
    component: NewsletterPage,
  },

  {
    path: "/newsletters/search",
    component: SearchNewslettersPage,
  },

  {
    path: "/newsletters/create",
    component: CreateNewsletter,
  },
];
