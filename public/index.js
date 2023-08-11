let longUrlError
let outputEl
let outputContainer
let longUrlInput


window.onload = () => {
    getSubmitButton()
    longUrlError = document.getElementById("longUrlError")
    outputEl = document.getElementById("shortUrlOutput")
    outputContainer = document.getElementById("shortUrlContainer")
}

async function getSubmitButton() {
    document.getElementById("submitButton").addEventListener("click", async e => {
        e.preventDefault()

        longUrlInput = document.getElementById("longUrl")

        if (longUrlInput.value == "") {
            showError("Enter a URL")
            console.log("Long url value: ", longUrlInput.value)
        }
        const postBody = {
            "longUrl": longUrlInput.value
        }
        const res = await fetch("/shorten", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(postBody)
        })

        if (res.ok) {
            const shortUrl = await res.json()
            showShortUrl(shortUrl)

        } else {
            const error = await res.json()
            showError(error)
        }

    })
}

function showShortUrl(shortUrl) {
    outputEl.setAttribute("href", shortUrl.shortUrl)
    outputEl.textContent = shortUrl.shortUrl
    outputContainer.classList.add("is-visible")
    longUrlInput.classList.remove("error")
    longUrlError.classList.remove("is-visible")
}

function showError(error) {
    outputEl.removeAttribute("href")
    outputContainer.classList.remove("is-visible")
    longUrlInput.classList.add("error")
    longUrlError.textContent = `${error.Error}`
    longUrlError.classList.add("is-visible")
}