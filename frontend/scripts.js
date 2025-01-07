document.addEventListener("DOMContentLoaded", () => {
    const today = new Date().toISOString().split("T")[0];
    const lastWeek = new Date();
    lastWeek.setDate(lastWeek.getDate() - 7);
    const sevenDaysAgo = lastWeek.toISOString().split("T")[0];

    document.getElementById("dateTo").value = today;
    document.getElementById("dateFrom").value = sevenDaysAgo;
});

document.querySelector("form").onsubmit = async (e) => {
    e.preventDefault();
    const overlay = document.getElementById("overlay");
    overlay.style.display = "block";
    const loader = document.getElementById("loader");
    loader.style.display = "block";
    const form = new FormData(e.target);
    const response = await fetch("/generate", {
        method: "POST",
        body: form
    });
    const data = await response.json();

    const qrCodeList = document.getElementById("qrCodeList");
    const transformedReceipts = document.getElementById("transformedReceipts");
    const saveJsonButton = document.getElementById("saveJson");
    const copyJsonButton = document.getElementById("copyJson");
    const qrCodesContainer = document.getElementById("qrCodes");

    if (data.qrCodes && data.qrCodes.length > 0) {
        qrCodeList.innerHTML = data.qrCodes.map(qr => `
            <div class="qrCodeItem">
                <img src="${qr.image}" alt="QR Code" onclick="copyToClipboard('${qr.text}')">
            </div>
        `).join("");
        transformedReceipts.textContent = data.transformedReceipts;
        transformedReceipts.style.display = "block";
        qrCodesContainer.style.display = "block";
        saveJsonButton.style.display = "inline-block";
        copyJsonButton.style.display = "inline-block";
    } else {
        qrCodeList.innerHTML = "";
        transformedReceipts.textContent = "";
        transformedReceipts.style.display = "none";
        qrCodesContainer.style.display = "none";
        saveJsonButton.style.display = "none";
        copyJsonButton.style.display = "none";
        showToast("No receipts found or loaded");
    }

    loader.style.display = "none";
    overlay.style.display = "none";
};

// Функция для копирования текста в буфер обмена
function copyToClipboard(text) {
    const textarea = document.createElement("textarea");
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    showToast("Text copied to clipboard");
}

// Функция для показа всплывающего окна
function showToast(message) {
    const toast = document.getElementById("toast");
    toast.textContent = message;
    toast.className = "toast show";
    setTimeout(() => {
        toast.className = toast.className.replace("show", "");
    }, 3000);
}

// Функция для сохранения содержимого textarea в файл
document.getElementById("saveJson").addEventListener("click", () => {
    const textarea = document.getElementById("transformedReceipts");
    const dateFrom = document.getElementById("dateFrom").value;
    const dateTo = document.getElementById("dateTo").value;
    const fileName = `${dateFrom.replace(/-/g, "")}-${dateTo.replace(/-/g, "")}.json`;

    const blob = new Blob([textarea.value], { type: "application/json" });
    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    showToast("JSON file saved");
});

// Функция для копирования содержимого textarea в буфер обмена
document.getElementById("copyJson").addEventListener("click", () => {
    const textarea = document.getElementById("transformedReceipts");
    copyToClipboard(textarea.value);
});
