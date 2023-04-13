
// initialize
getDbs()

// getPage
function getPage(num) {
    var collection = document.getElementById("collections").value
	if (collection == "Select Collection") {
		return
	}
	page = Number(document.getElementById("page").value)
	if (page <= 1 && num == -1) {
		return
	}
	page += num

	getDocuments(collection, page)
}

function find() {
    var collection = document.getElementById("collections").value
	if (collection == "Select Collection") {
		return
	}
	getDocuments(collection, 1)
}

// getDbs request
function getDbs() {
    var list = document.getElementById("dbs")

    fetch("/dbs", {
        method: "GET"
    }).then((res) => {
        return res.json()
    }).then((res) => {
        res.items.forEach(item => {
            var option = document.createElement('option')
            option.value = item
            option.innerHTML = item
            list.appendChild(option)
        })
    }).catch((reason) => {
        var list = document.getElementById("documents")
        list.innerHTML = reason
    })
}

// getCollections request
function getCollections(db) {
    var list = document.getElementById("collections")
    list.innerHTML = "<option selected>Select Collection</option>"

    fetch("/dbs/" + db + "/collections", {
        method: "GET"
    }).then((res) => {
        return res.json()
    }).then((res) => {
        res.items.forEach(item => {
            var option = document.createElement('option')
            option.value = item
            option.innerHTML = item
            list.appendChild(option)
        })
    }).catch((reason) => {
        var list = document.getElementById("documents")
        list.innerHTML = reason
    })
}

// getDocuments request
function getDocuments(collection, page) {
	document.getElementById("page").value = page
    var list = document.getElementById("documents")
    list.innerHTML = ""
    var db = document.getElementById("dbs").value
    var head = document.createElement("thead")
    var body = document.createElement("tbody")
    var columns = []
    var limit = 50

    var key = document.getElementById("key").value
    var value = document.getElementById("value").value

    // documents
    fetch("/dbs/" + db + "/collections/" + collection + "/documents?page=" + page + "&limit=" + limit + "&key=" + key + "&value=" + value, {
        method: "GET"
    }).then((res) => {
        return res.json()
    }).then((res) => {
        // columns
        var tr = document.createElement("tr")
        res.items.forEach(item => {
            Object.keys(item).forEach(key => {
                if (!columns.includes(key)) {
                    columns.push(key)
                }
            })
        })
        columns.forEach(col => {
            var th = document.createElement("th")
            th.scope = "col"
            th.innerHTML = col
            tr.appendChild(th)
        })
        head.appendChild(tr)
        list.appendChild(head)

        // rows
        res.items.forEach(item => {
            var tr = document.createElement("tr")
            columns.forEach(col => {
                var td = document.createElement("td")
                Object.keys(item).forEach(key => {
                    if (key == col) {
                        td.innerHTML = JSON.stringify(item[key])
                    }
                })
                tr.appendChild(td)
            })
            body.appendChild(tr)
        })
        list.appendChild(body)
    }).catch((reason) => {
        list.innerHTML = reason
    })
}
