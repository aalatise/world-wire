import { Observable } from '@firebase/util';
import { IInstitution } from '../../shared/models/participant.interface';
import { ActivatedRouteSnapshot, Resolve } from '@angular/router';
import { Injectable } from '@angular/core';
import { SessionService } from '../services/session.service';
import * as _ from 'lodash';
import { AuthService } from '../services/auth.service';
import { PortalApiService } from '../services/portal-api.service';
import { get } from 'lodash';

@Injectable()
export class ParticipantResolve implements Resolve<string> {

  constructor(
    private authService: AuthService,
    private sessionService: SessionService,
    private portalApiService: PortalApiService,
  ) {
  }

  resolve(
    route: ActivatedRouteSnapshot,
    // state: RouterStateSnapshot
  ): Observable<any> | Promise<any> | any {

    // resolve with participant details

    return new Promise(async (resolve) => {

      const getPreviousSlug: string =
              _.has(this.sessionService, 'institution.info.slug') ? this.sessionService.institution.info.slug : '';

      // check if current route matches the previously saved participant in session.service.ts
      if (getPreviousSlug === route.params.slug) {

        // no need to query institution if it is already stored in the service
        resolve(this.sessionService.institution);

      } else {
        const slug: string = route.paramMap.get('slug');

        if (get(this.sessionService, 'institution.info.slug') === slug) {
          resolve(this.sessionService.institution);
        }

        this.portalApiService.getInstitution(slug).subscribe((institution) => {
          this.sessionService.institution = institution;
          resolve(institution);
        }, error => {
          resolve(null);
        });
      }
    });
  }
}
