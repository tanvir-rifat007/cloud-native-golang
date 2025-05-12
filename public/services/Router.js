import { routes } from "./Routes.js";

export const Router = {
  init: () => {
    document.querySelectorAll("a").forEach((a) => {
      a.addEventListener("click", (event) => {
        event.preventDefault();

        const href = a.getAttribute("href");
        Router.go(href);
      });
    });
    window.addEventListener("popstate", () => {
      Router.go(location.pathname, false);
    });
    // Process initial URL
    console.log(location.pathname);
    Router.go(location.pathname + location.search);
  },
  go: (route, addToHistory = true) => {
    document.querySelectorAll("a").forEach((link) => {
      const linkHref = link.getAttribute("href");
      if (linkHref === route) {
        link.classList.add(
          "text-indigo-300",
          "font-semibold",
          "border-b-2",
          "border-indigo-500"
        );
      } else {
        link.classList.remove(
          "text-indigo-300",
          "font-semibold",
          "border-b-2",
          "border-indigo-500"
        );
      }
    });

    if (addToHistory) {
      history.pushState(null, "", route);
    }
    const routePath = route.includes("?") ? route.split("?")[0] : route;
    let pageElement = null;
    let needsLogin;
    for (const r of routes) {
      console.log(route);
      if (typeof r.path === "string" && r.path === routePath) {
        pageElement = new r.component();
        needsLogin = r.loggedIn === true;
        break;
      } else if (r.path instanceof RegExp) {
        const match = r.path.exec(route);
        console.log(match);
        if (match) {
          const params = match[1];
          pageElement = new r.component();
          pageElement.params = params;
          needsLogin = r.loggedIn === true;

          break;
        }
      }
    }

    if (pageElement) {
      // A page was found, we checked if we have access to it.
      if (needsLogin && app.store.loggedIn == false && !app.store.activated) {
        app.router.go("/account/login");
        return;
      }
    }

    if (pageElement == null) {
      pageElement = document.createElement("h1");
      pageElement.textContent = "Page not found";
    }

    function changePage() {
      document.querySelector("main").innerHTML = "";
      document.querySelector("main").appendChild(pageElement);
    }

    if (!document.startViewTransition) {
      changePage();
      return;
    }
    document.startViewTransition(() => {
      changePage();
    });
  },
};

export default Router;
