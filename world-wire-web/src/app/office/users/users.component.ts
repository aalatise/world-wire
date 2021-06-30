import { Component, OnInit, OnDestroy, NgZone } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { IRolesOptions, IUserSuperPermissions, IRoles, IUserProfile } from '../../shared/models/user.interface';
import { SuperPermissionsService } from '../../shared/services/super-permissions.service';
import {
  ISuperPermissionsDialogData,
  SuperPermissionsDialogComponent
} from '../../shared/components/super-permissions-dialog/super-permissions-dialog.component';
import { AuthService } from '../../shared/services/auth.service';
import { Observable, Subscription } from 'rxjs';
import { KeyValue } from '@angular/common/src/pipes/keyvalue_pipe';

@Component({
  templateUrl: './users.component.html',
  styleUrls: ['./users.component.scss']
})
export class UsersComponent implements OnInit, OnDestroy {

  users: IUserProfile[];
  disable: any;
  humanizeRoles: any = '';

  userSub: Subscription;

  // emailSortDesc = (a: KeyValue<string, IRoles>, b: KeyValue<string, IRoles>) => {
  //   const emailA = a.value.email.toLowerCase();
  //   const emailB = b.value.email.toLowerCase();

  //   if (emailA > emailB) {
  //     return 1;
  //   } else if (emailA < emailB) {
  //     return -1;
  //   }
  //   return 0;
  // }

  constructor(
    public dialog: MatDialog,
    public superPermissionsService: SuperPermissionsService,
    public authService: AuthService,
    private ngZone: NgZone
  ) {
    this.disable = this.superPermissionsService.disable;
    this.humanizeRoles = this.superPermissionsService.humanizeRoles;
  }

  ngOnInit() {
    this.getUsers();
  }

  ngOnDestroy() {
    this.userSub.unsubscribe();
  }

  /**
   * add user button click to open modal
   *
   * @memberof UsersComponent
   */
  addUser() {
    this.openUserDialog('add');
  }

  /**
   * add user button click to open modal
   *
   * @memberof UsersComponent
   */
  editUser(email: string, userId: string, role: IRolesOptions) {
    this.openUserDialog('edit', email, userId, role);
  }

  deleteUser(email: string, userId: string, role: IRolesOptions) {
    this.openUserDialog('remove', email, userId, role);
  }

  /**
   * open the user dialog for edit and save
   *
   * @private
   * @param {('add' | 'edit')} action
   * @param {string} [email] used to 'remove' or 'edit'
   * @param {string} [userId] used to 'remove' or 'edit'
   * @memberof UsersComponent
   */
  private openUserDialog(action: 'add' | 'edit' | 'remove', email?: string, userId?: string, role?: IRolesOptions) {

    const data: ISuperPermissionsDialogData = {
      action: action,
      email: email,
      userId: userId,
      role: role
    };

    const dialogRef = this.dialog.open(SuperPermissionsDialogComponent, {
      disableClose: true,
      data: data
    });

    dialogRef.afterClosed().subscribe(result => this.getUsers());

  }

  private getUsers() {
    this.userSub = this.superPermissionsService.getUsersWithSuperPermission().subscribe(users => this.users = users);
  }

}
