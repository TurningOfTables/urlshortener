window.onload = () => {
    getSubmitButton()
}

async function getSubmitButton() {
    document.getElementById("submitButton").addEventListener("click", async e => {
        e.preventDefault()
        const url = document.getElementById("longUrl").value
        const postBody = {
            "longUrl": url
        }
        const res = await fetch("/shorten", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(postBody)
        })

        const outputEl = document.getElementById("shortUrlOutput")
        if (res.ok) {
            const shortUrl = await res.json()
            outputEl.setAttribute("href", shortUrl)
            outputEl.textContent = shortUrl
        } else {
            outputEl.removeAttribute("href")
            outputEl.textContent = "Error :("
        }
    })
}