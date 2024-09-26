function main(input) {
    var inData = JSON.parse(input.Data);
    var outData = {}
    
    outData["public"] = {"name": inData.name, "occupation": inData.occupation}
    outData["private"] = {"abilities": inData.abilities, "abilities": inData.abilities, "collaborators": inData.collaborators}
    
    return {
        Data: outData,
    }
}

