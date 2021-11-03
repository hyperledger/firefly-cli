
	const Web3 = require('web3')
	const fs = require('fs')
	
	function generateKeyFile(privateKey, memberName) {
		console.log("Inside keygen")
		const web3 = new Web3(new Web3.providers.HttpProvider('http://127.0.0.1:8555'))
		var V3KeyStore = web3.eth.accounts.encrypt(privateKey, "SomeSüper$trÖngPäs$worD!");
		var KeyFile = JSON.stringify(V3KeyStore)
		console.log("Keyfile: ", KeyFile)
		// return KeyFile;
		fs.writeFile(memberName, KeyFile, function (err) {
			if (err) {
				console.log("Error: ", err.message)
			} else {
				console.log("Success")
			}
		})
	}

	const dirPath = "/usr/local/bin/accounts"
	fs.mkdirSync("/usr/local/bin" + "/keyFiles", { recursive: true })
  	files = fs.readdirSync(dirPath)
	console.log("Files: ", files)
	for (var i=0; i < files.length; i++) {
		pass_file = dirPath + "/" + files[i] + "/privateKey"
		mem_name = "/usr/local/bin" + "/keyFiles/" + files[i] + "_keyFile"
		try {
			var data = fs.readFileSync(pass_file, "utf-8") 
			generateKeyFile(String(data), mem_name)
		} catch (err) {
			console.error(err)
		}

	}
	


	