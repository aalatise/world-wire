import { Injectable } from '@angular/core';
import { HttpEvent, HttpInterceptor, HttpHandler, HttpRequest, HttpHeaders } from '@angular/common/http';

import { Observable } from 'rxjs';
import { SessionService } from '../services/session.service';

@Injectable()
export class AuthenticationInterceptor implements HttpInterceptor {
    constructor(private session: SessionService) {}
    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (this.session.accessToken) {
            req = req.clone({
                headers: new HttpHeaders({
                    'Authorization': 'Bearer ' + localStorage.getItem('id_token')
                })
            })
        }
        return next.handle(req);
    }
}