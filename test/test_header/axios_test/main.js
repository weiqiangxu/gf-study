import Vue from 'vue'
import axios from 'axios'
import VueAxios from 'vue-axios'

Vue.use(VueAxios, axios)

Vue.axios.get("http://localhost/api/test").then((response) => {
  console.log(response.data)
})

axios({
 	headers: {
        // application/json ： 请求体中的数据会以json字符串的形式发送到后端
        // application/x-www-form-urlencoded：请求体中的数据会以普通表单形式（键值对）发送到后端
        // multipart/form-data： 它会将请求体的数据处理为一条消息，以标签为单元，用分隔符分开。既可以上传键值对，也可以上传文件。
        'Content-Type': 'application/x-www-form-urlencoded' //参数为object时候请求体中的数据会以普通表单形式（键值对）发送到后端
    },
  method: 'post',
  url: 'http://localhost/api/test',
  params: {
    name: 'jack'
  }
});