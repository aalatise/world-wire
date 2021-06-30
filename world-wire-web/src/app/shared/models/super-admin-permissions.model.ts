import * as _ from 'lodash';
import { Injectable } from '@angular/core';
import { PortalApiService } from '../services/portal-api.service';
import { catchError } from 'rxjs/operators';
import { of } from 'rxjs';

@Injectable()
export class SuperAdminPermissionsModel {

    /**
     * Used to determine if a user is a super admin
     * Security: Is a SECURE node for permissions
     * @memberof SuperAdminPermissionsModel
     */
    route = 'super_permissions/{userId}/role';

    constructor(
        private portalApiService: PortalApiService
    ) { }

    /**
     * Add user permissions.
     * NOTE: permissions are added and removed one at at time.
     *
     * @returns {Promise<void>}
     * @memberof SuperAdminPermissionsModel
     */
    add(
        userId: string,
        roleType: 'admin' | 'manager'
    ): Promise<void> {
        // return promise since result is a single success or failure
        return new Promise((resolve, reject) => {
          this.portalApiService.updateSuperPermission({
            userId: userId,
            role: roleType,
          }).pipe(
            catchError(err => of([])),
          ).subscribe(() => {
            resolve();
          }, (error) => {
            console.log('Error: Unable to add super admin permissions.', error);
            alert('Error: Unable to add super admin permissions.');
            reject();
          });
        });
    }

    /**
     * Remove user permissions.
     * NOTE: permissions are added and removed one at at time.
     *
     * @returns {Promise<void>}
     * @memberof SuperAdminPermissionsModel
     */
    remove(
        userId: string,
        roleType: 'admin' | 'manager'
    ): Promise<void> {

        // return promise since result is a single success or failure
        return new Promise((resolve, reject) => {
          this.portalApiService.deleteSuperPermission(userId).pipe(
            catchError(err => of([])),
          ).subscribe(() => {
            resolve();
          }, (error) => {
            console.log('Error: Unable to remove super admin permissions.', error);
            alert('Error: Unable to remove super admin permissions.');
            reject();
          });
        });
    }
}


