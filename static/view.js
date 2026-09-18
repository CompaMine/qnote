(function () {
  const ready = document.getElementById("ready");
  const content = document.getElementById("content");
  const gone = document.getElementById("gone");
  const revealBtn = document.getElementById("reveal-btn");
  const plaintextEl = document.getElementById("plaintext");
  const errorEl = document.getElementById("error");

  const id = location.pathname.split("/").pop();
  const keyHex = location.hash.slice(1);

  function show(el) {
    [ready, content, gone].forEach((node) => {
      node.classList.add("hidden");
    });
    errorEl.classList.add("hidden");
    el.classList.remove("hidden");
    el.style.animation = "none";
    void el.offsetHeight;
    el.style.animation = "";
  }

  function setLoading(btn, on) {
    btn.disabled = on;
    btn.classList.toggle("loading", on);
  }

  if (!id || !keyHex) {
    show(gone);
  }

  revealBtn.addEventListener("click", async () => {
    if (!id || !keyHex) {
      show(gone);
      return;
    }

    setLoading(revealBtn, true);
    try {
      const key = await QNotesCrypto.importKeyFromHex(keyHex);

      const res = await fetch(`/api/notes/${encodeURIComponent(id)}`);
      if (res.status === 404) {
        show(gone);
        return;
      }
      if (!res.ok) {
        throw new Error("не удалось получить записку");
      }

      const data = await res.json();
      const text = await QNotesCrypto.decryptText(data.ciphertext, key);
      plaintextEl.textContent = text;
      show(content);

      history.replaceState(null, "", location.pathname);
    } catch (err) {
      if (err.name === "OperationError") {
        show(gone);
        return;
      }
      errorEl.textContent = err.message || "Ошибка расшифровки";
      errorEl.classList.remove("hidden");
      ready.classList.add("hidden");
    } finally {
      setLoading(revealBtn, false);
    }
  });
})();
