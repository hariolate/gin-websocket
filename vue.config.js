const path = require("path");

module.exports = {
    outputDir: path.resolve(__dirname, "./static"),
    publicPath: __dirname,
    configureWebpack: {
        resolve: {
            alias: {
                vue$:'vue/dist/vue.esm.js'
            }
        }
    }
};