import * as _ from 'lodash';
import { ROLES } from '../constants/general.constants';
import { Injectable, NgZone, isDevMode } from '@angular/core';
import { HttpClient, HttpHeaders, HttpRequest, HttpParams } from '@angular/common/http';
import { environment } from '../../../environments/environment';
import { AuthService } from './auth.service';
import { IUserParticipantPermissions, IUserSuperPermissions, IUserProfile } from '../models/user.interface';
import { Confirm2faService } from './confirm2fa.service';
import { ParticipantPermissionsService } from './participant-permissions.service';
import { Observable, Observer } from 'rxjs';
import { PortalApiService } from './portal-api.service';
import { SessionService } from './session.service';

@Injectable()
export class SuperPermissionsService {

    roles = ROLES;

    constructor(
        private http: HttpClient,
        private authService: AuthService,
        private confirm2Fa: Confirm2faService,
        private participantPermissionsService: ParticipantPermissionsService,
        private portalApiService: PortalApiService,
        private sessionService: SessionService,
    ) {
        // Do NOT create a model here rather pass the model into the angular
        // service method call. This makes it possible to use the methods as
        // service and models as objects
    }

    /**
     * disables button in view so that users cannot
     * edit or change their own permissions
     *
     * @param {string} email1
     * @param {string} email2
     * @returns
     * @memberof UsersComponent
     */
    disable = this.participantPermissionsService.disable;

    /**
     * formats permissions array to human readable string in view
     *
     * @param {IRolesOptions} roles
     * @returns
     * @memberof UsersComponent
     */
    humanizeRoles = this.participantPermissionsService.humanizeRoles;

    /**
     * Get all super users
     *
     * @returns {Observable<IUserSuperPermissions>}
     * @memberof SuperPermissionsService
     */
    getUsersWithSuperPermission(): Observable<IUserProfile[]> {
        return this.portalApiService.getSuperPermissions();
    }

    /**
     * Creates new (or updates) user permissions and returns uid
     * (or returns existing user uid for email)
     * Alias: SuperPermissionsService.setUserId()
     *
     * @param {string} institutionId
     * @param {('admin' | 'manager' | 'viewer')} role
     * @param {string} email
     * @returns
     * @memberof ParticipantPermissionsService
     */
    update(role: 'admin' | 'manager' | 'viewer', email: string) {
        return new Promise((resolve, reject) => {

            this.authService.getFirebaseIdToken().then((h: HttpHeaders) => {

                // update user's permissions

                // validate required fields are present
                if (!_.isEmpty(email) && !_.isEmpty(role)) {

                    // update permissions node in firebase
                    // NOTE: doing as an http post instead of calling firebase db
                    // directly because this requires the success of two separate calls to
                    // firebase (in the case of adding a user -> one call to create the user,
                    // and one call to add their permissions). Creating a http post request
                    // ensures graceful failure
                    h = h.set('Authorization', `Bearer ${this.sessionService.accessToken}`);
                    h = h.set('Accept', 'text');

                    const r = new HttpRequest(
                        'POST',
                        environment.apiRootUrl + '/permissions/super',
                        {
                            email: email,
                            role: role
                        },
                        { headers: h, responseType: 'text'}
                    );

                    this.confirm2Fa.go(r)
                        .then((uid: string) => {
                            resolve(uid);
                        }, (err) => {
                            if (isDevMode()) {
                                console.log('Error: Unable to add super permissions.', err);
                            }
                            reject();
                        });

                }

            });

        });

    }


    /**
     * Remove user permissions.
     * NOTE: permissions are added and removed one at at time.
     *
     * @param {string} userId
     * @returns {Promise<void>}
     * @memberof SuperPermissionsService
     */
    remove(userId: string): Promise<void> {

        // validate required fields are present
        if (!_.isEmpty(userId)) {

            // TODO: create a DELETE endpoint in auth-service to handle this instead, similar to update()
            // return promise since result is a single success or failure
            return new Promise(async (resolve, reject) => {

                try {
                    await this.portalApiService.deleteSuperPermissionP(userId);
                    resolve();
                } catch (err) {
                    reject(err);
                }
            });
        }
    }
}
