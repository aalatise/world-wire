import { Injectable } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, Router, RouterStateSnapshot } from '@angular/router';
import { SessionService } from '../services/session.service';
import { get, isEmpty } from 'lodash';
import { AuthService } from '../services/auth.service';
import { IInstitution } from '../models/participant.interface';
import { PortalApiService } from '../services/portal-api.service';

/***
 * ParticipantPermissionsGuard
 * protects routes that require participant permissions
 * or super permissions to view
 */
@Injectable()
export class ParticipantPermissionsGuard implements CanActivate {

    // store slug from route params to use across functions
    slug: string;

    constructor(
        private authService: AuthService,
        private sessionService: SessionService,
        private router: Router,
        private portalApiService: PortalApiService,
    ) { }

    async canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ) {

        try {

            // get slug of parent route to activate child route
            const url: string[] = state.url.split('/');

            this.slug = url.length > 2 ? url[2] : '';

            // set userProfile on authService
            const userProfile = await this.authService.getUserProfile(this.authService);

            // set institution on session
            const selectedInstitution = await this.getInstitution();

            // logical redirect based on if:
            // 1) user has permissions
            // 2) institution is available in the session
            // check if user profile exists otherwise redirect to login
            if (isEmpty(userProfile)) {

                // unable to determine if the user has permissions for this institution
                // redirect user to login
                this.router.navigate(['/login']);
                return false;

            } else {

                // check if institution exists otherwise redirect to not-found
                if (isEmpty(selectedInstitution)) {

                    // unable to set the institution
                    // redirect to unauthorized page
                    this.router.navigate(['/not-found']);
                    return false;

                } else {

                    // check if user has access rights to proceed to route based
                    // on specified permissions on portals 'route.data' in portal-routing.ts
                    if (
                        // check if user has "PARTICIPANT" specific permissions to access route
                        this.authService.hasParticipantPermissions(
                            userProfile,
                            selectedInstitution.info.institutionId,
                            route.data.participant_permissions) === true ||
                        // or
                        // check if user has "SUPER" permissions to access route
                        this.authService.hasSpecificSuperPermissions(userProfile, route.data.super_permissions) === true
                    ) {
                        // success resolve true and allow transition
                        // activate route since permissions exist
                        return true;
                    } else {
                        // unable to determine if the user has permissions for this institution
                        // redirect to unauthorized page
                        this.router.navigate(['/unauthorized']);
                        return false;
                    }

                }

            }

        } catch (error) {
            // most likely one of the promises related
            // to getting the institution or user permissions failed
            console.log(error);
            this.router.navigate(['/not-found']);
            return false;
        }

    }

    /**
     * checks if the session already includes the institution
     * associated with the selected routes current slug,
     * otherwise it retrieves the institution information from
     * the db and sets it to the session based
     * on the specified route's slug
     *
     * @private
     * @returns {Promise<IInstitution>}
     * @memberof ParticipantPermissionsGuard
     */
    private getInstitution(): Promise<IInstitution> {

      return new Promise(async (resolve) => {
        if (get(this.sessionService, 'institution.info.slug') === this.slug) {
          resolve(this.sessionService.institution);
        }

        this.portalApiService.getInstitution(this.slug).subscribe((institution) => {
          this.sessionService.institution = institution;
          resolve(institution);
        }, error => {
          resolve(null);
        });
      });
    }
}
