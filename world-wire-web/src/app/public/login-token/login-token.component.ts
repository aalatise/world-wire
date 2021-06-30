import { Component, OnInit, HostBinding, NgZone, isDevMode } from '@angular/core';
import { AuthService } from '../../shared/services/auth.service';
import { ActivatedRoute, Router } from '@angular/router';
import { IUserProfile } from '../../shared/models/user.interface';
import { get, isEmpty } from 'lodash';

@Component({
  selector: 'app-login',
  templateUrl: './login-token.component.html',
  styleUrls: ['./login-token.component.scss'],
})
export class LoginTokenComponent implements OnInit {

  constructor(
    public authService: AuthService,
    private route: ActivatedRoute,
    private ngZone: NgZone,
    private router: Router,
  ) {
  }

  @HostBinding('attr.class') cls = 'flex-fill';

  async ngOnInit() {

    // NOTE: no need to put this in a route guard or parent component
    // because this is only called after IBMId login callback is
    // fired from back-end node.js server.

    // check if firebase user exists
    if (isEmpty(this.authService.userProfile)) {

      // get data from route params
      const data = this.route.snapshot.queryParams as { token: string };

      // if token is provided in query param
      // try to use token to login to firebase
      if (data.token) {

        try {
          const userProfile = await this.authService.getUserProfile(this.authService);
          this.redirect(userProfile);
        } catch (err) {
          if (isDevMode()) {
            console.log(err);
          }
          this.router.navigate(['/unrecognized']);
        }

      } else {

        // redirect user to login via IBM Id
        this.authService.signInIbmId();

      }

    } else {

      // already logged in so redirect
      const userProfile = await this.authService.getUserProfile();
      this.redirect(userProfile);

    }

  }

  /**
   * redirect user to where they have permissions
   *
   * @param {IUser} userProfile
   * @memberof LoginTokenComponent
   */
  redirect(userProfile: IUserProfile) {

    const hasSuperPermission = Object.keys(userProfile.super_permission.roles).some((key) => userProfile.super_permission.roles[key]);
    if (userProfile.super_permission && hasSuperPermission) {
      this.ngZone.run(() => {
        return this.router.navigate(['/office']);
      });
    } else if (userProfile.participant_permissions) {
      const hasParticipantPermission = Object.keys(userProfile.participant_permissions).some((key) => {
        const hasValidPermission = Object.keys(userProfile.participant_permissions[key].roles).some((roleKey) => {
          return userProfile.participant_permissions[key].roles[roleKey];
        });
        return hasValidPermission;
      });
      if (hasParticipantPermission) {
        this.ngZone.run(() => {
          return this.router.navigate(['/portal']);
        });
      }
    } else {
      this.ngZone.run(() => {
        return this.router.navigate(['/unrecognized']);
      });
    }
  }
}
