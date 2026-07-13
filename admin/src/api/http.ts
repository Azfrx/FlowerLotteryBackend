import axios from 'axios';
export const http=axios.create({baseURL:'/api/v1/admin',timeout:15000});
http.interceptors.request.use(config=>{const token=localStorage.getItem('admin_access_token');if(token)config.headers.Authorization=`Bearer ${token}`;return config});
http.interceptors.response.use(response=>response.data,error=>Promise.reject(error));
