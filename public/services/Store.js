const Store = {
  token: localStorage.getItem("token") || null,
};

const proxiedStore = new Proxy(Store, {
  set(target, prop, value) {
    if (prop === "token") {
      localStorage.setItem("token", value);
    }
    target[prop] = value;
    return true;
  },
});

export default proxiedStore;
