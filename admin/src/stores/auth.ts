import {defineStore} from 'pinia'; import {ref} from 'vue';
export const useAuthStore=defineStore('auth',()=>{const token=ref(localStorage.getItem('admin_access_token')||'');function setToken(value:string){token.value=value;localStorage.setItem('admin_access_token',value)}function logout(){token.value='';localStorage.removeItem('admin_access_token')}return{token,setToken,logout}})
