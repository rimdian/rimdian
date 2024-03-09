const CracoLessPlugin = require('craco-less')

module.exports = {
  plugins: [
    {
      plugin: CracoLessPlugin,
      options: {
        lessLoaderOptions: {
          lessOptions: {
            modifyVars: {
              // '@primary-color': '#1DA57A'
            },
            javascriptEnabled: true
          }
        }
      }
    }
  ],
  babel: {
    plugins: [
      [
        'prismjs',
        {
          languages: ['javascript', 'css', 'markup', 'json', 'sql'],
          plugins: ['line-numbers'],
          theme: 'default',
          css: true
        }
      ]
    ]
  }
}
