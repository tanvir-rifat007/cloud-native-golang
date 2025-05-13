const Store = {
  token: localStorage.getItem("token") || null,
  createdBy: localStorage.getItem("createdBy") || null,
};

const proxiedStore = new Proxy(Store, {
  set(target, prop, value) {
    if (prop === "token") {
      localStorage.setItem("token", value);
    }

    if (prop === "createdBy") {
      localStorage.setItem("createdBy", value);
    }

    target[prop] = value;
    return true;
  },
});

export default proxiedStore;
