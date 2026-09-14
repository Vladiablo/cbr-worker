import http from 'k6/http';
// import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    vus: 10,
    duration: '30s'
};

export default () => {
    http.get(`http://127.0.0.1:8080/v1/getRates`);
};
