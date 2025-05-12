export const API = {
  baseURL: "http://localhost:8080/api/v1",

  postNewsletter: async (body) => {
    return API.fetchData("/newsletter", body);
  },

  postNewsletters: async (body) => {
    return API.fetchData("/newsletter/create", body);
  },
  getNewsletters: async () => {
    return API.fetchNewsletters("/newsletters");
  },
  getNewsletter: async (id) => {
    return API.fetchNewsletters(`/newsletter/${id}`);
  },

  searchNewsletters: async (query) => {
    return API.fetchNewsletters("/newsletters/search", query);
  },

  fetchNewsletters: async (url, args) => {
    try {
      const searchParams = args ? new URLSearchParams(args).toString() : "";
      const response = await fetch(API.baseURL + url + "?" + searchParams, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      const res = await response.json();

      return res;
    } catch (error) {
      console.error("Error fetching data:", error);
      throw error;
    }
  },

  fetchData: async (url, data = {}) => {
    try {
      const response = await fetch(API.baseURL + url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const res = await response.json();

      return res;
    } catch (error) {
      console.error("Error fetching data:", error);
      throw error;
    }
  },
};
