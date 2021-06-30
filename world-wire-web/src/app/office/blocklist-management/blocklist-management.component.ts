import { Component, OnInit, HostBinding, NgZone, OnDestroy } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { BlocklistService } from './../../shared/services/blocklist.service';
import { BlocklistDialogComponent } from './blocklist-dialog/blocklist-dialog.component';
import { find } from 'lodash';
import { ApprovalInfo } from '../../shared/models/approval.interface';
import { BlocklistType, BlocklistRequest } from '../../shared/models/blocklist.interface';
import { HttpErrorResponse, HttpClient, HttpParams } from '@angular/common/http';
import { NotificationService } from '../../shared/services/notification.service';
import { SuperApprovalsModel } from '../../shared/models/super-approval.model';
import { WorldWireError } from '../../shared/models/error.interface';
import { AuthService } from '../../shared/services/auth.service';
import * as cc from 'currency-codes';
import * as countries from 'i18n-iso-countries';
import { environment } from '../../../environments/environment';
import { Subscription } from 'rxjs';
import { PortalApiService } from '../../shared/services/portal-api.service';

export interface BlocklistItem {
  name: string;
  isoCode: string;
  exists: boolean;
  status: 'pending' | 'approved';
  request?: BlocklistRequest;
  approval?: ApprovalInfo;
}

export interface Currency extends BlocklistItem {
  alternateName?: string;
}
export interface Country extends BlocklistItem {
  alternateName?: string;
}

@Component({
  selector: 'app-blocklist-management',
  templateUrl: './blocklist-management.component.html',
  styleUrls: ['./blocklist-management.component.scss'],
  providers: [
    BlocklistService,
    SuperApprovalsModel
  ]
})

export class BlocklistManagementComponent implements OnInit, OnDestroy {

  @HostBinding('attr.class') cls = 'flex-fill';

  // holds list of blocked currencies
  public currencies: Currency[];

  currenciesLoaded = false;

  // holds list of blocked countries
  public countries: Country[];

  countriesLoaded = false;

  private subscriptions: Subscription[];

  loadingRequest = false;

  public constructor(
    public dialog: MatDialog,
    public authService: AuthService,
    public blocklistService: BlocklistService,
    private notificationService: NotificationService,
    private superApprovals: SuperApprovalsModel,
    private http: HttpClient,
    private portalApiService: PortalApiService,
  ) { }

  public ngOnInit() {
    this.subscriptions = [];
    countries.registerLocale(require(`i18n-iso-countries/langs/${this.authService.userLangShort}.json`));
    this.getBlockedList();
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }

  /**
   * Gets list of blocked countries and currencies
   *
   * @memberof BlocklistManagementComponent
   */
  public getBlockedList() {

    // reset loading
    this.countriesLoaded = false;
    this.currenciesLoaded = false;

    // empty lists
    this.countries = null;
    this.currencies = null;

    const countryPromises: Promise<any>[] = [
      this.blocklistService.getBlocklist('country'),
      this.blocklistService.getBlocklistRequests('country')
    ];
    // get list of blocked countries from API

    Promise.all(countryPromises).then(async (results) => {

      const blockedCountryCodes: string[] = results[0];

      const requests: BlocklistRequest[] = results[1];

      if (requests) {

        for (const request of requests) {

          if (request.approvalId) {

            const countryName: string = countries.getName(request.value, this.authService.userLangShort);

            const country: Country = {
              name: countryName,
              isoCode: request.value,
              exists: false,
              status: 'pending',
            };

            country.request = request;

            const latestAppproval: ApprovalInfo = await this.superApprovals.getApprovalInfo(request.approvalId);

            country.approval = latestAppproval;

            country.status = latestAppproval.requestApprovedBy ? 'approved' : 'pending';

            if (!this.countries) {
              this.countries = [];
            }

            this.countries.push(country);
          }
        }
      }

      if (blockedCountryCodes && blockedCountryCodes.length > 0) {
        this.countries = this.countries ? this.countries : [];

        for (const code of blockedCountryCodes) {

          const foundRequest = find(this.countries, (country: Country) => {
            return country.isoCode === code;
          });

          if (!foundRequest) {
            const countryName: string = countries.getName(code, this.authService.userLangShort);

            if (countryName) {
              this.countries.push({
                name: countryName,
                exists: true,
                isoCode: code,
                status: 'approved'
              });
            }
          } else {
            foundRequest.exists = true;
          }
        }
      }

      this.countriesLoaded = true;
    }).catch((err: HttpErrorResponse) => {

      const message = err.error ? err.message : 'Request could not be made to get the blocklist.';

      this.notificationService.show(
        'error',
        message,
        null,
        'Network Error',
        'top'
      );

      this.countriesLoaded = true;
    });

    const currencyPromises: Promise<any>[] = [
      this.blocklistService.getBlocklist('currency'),
      this.blocklistService.getBlocklistRequests('currency')
    ];

    // get list of blocked currencies from API
    Promise.all(currencyPromises).then(async (results) => {

      const blockedCurrencyCodes: string[] = results[0];

      const requests: BlocklistRequest[] = results[1];

      if (requests) {

        for (const request of requests) {

          if (request.approvalId) {

            const currencyName: string = cc.code(request.value).currency;

            const currency: Currency = {
              name: currencyName,
              isoCode: request.value,
              exists: false,
              status: 'pending',
            };

            currency.request = request;

            const latestAppproval: ApprovalInfo = await this.superApprovals.getApprovalInfo(request.approvalId);

            currency.approval = latestAppproval;

            currency.status = latestAppproval.requestApprovedBy ? 'approved' : 'pending';

            if (!this.currencies) {
              this.currencies = [];
            }

            this.currencies.push(currency);
          }
        }
      }

      if (blockedCurrencyCodes && blockedCurrencyCodes.length > 0) {
        this.currencies = this.currencies ? this.currencies : [];

        for (const code of blockedCurrencyCodes) {
          const currencyExists = cc.code(code);

          const foundCurrency = find(this.currencies, (currency: Currency) => {
            return currency.isoCode === code;
          });

          if (currencyExists && !foundCurrency) {
            this.currencies.push({
              name: currencyExists.currency,
              isoCode: currencyExists.code,
              exists: true,
              status: 'approved'
            });
          } else {
            foundCurrency.exists = true;
          }
        }
      }

      this.currenciesLoaded = true;
    }).catch((err: HttpErrorResponse) => {
      const message = err.error ? err.message : 'Request could not be made to get the blocklist.';

      this.notificationService.show(
        'error',
        message,
        null,
        'Network Error',
        'top'
      );

      this.currenciesLoaded = true;
    });

  }

  /**
   * Approve new blocklist request
   *
   * @param {BlocklistItem} item
   * @memberof BlocklistManagementComponent
   */
  public approveBlocklistRequest(item: BlocklistItem) {

    // Checker cannot be the same as maker
    if (item.approval.requestInitiatedBy === this.authService.userProfile.profile.email) {

      this.notificationService.show(
        'error',
        'Approver of request cannot be the same as the requestor.',
        null,
        'Unauthorized Action',
        'top'
      );
      return;
    }

    this.loadingRequest = true;

    this.blocklistService.addToBlocklist(item.request, 'approve').then((response: any) => {

      const deleteSub = this.portalApiService.deleteBlocklistRequest(item.approval.key).subscribe(() => {
        this.loadingRequest = false;
        this.getBlockedList();
      }, err => {
        this.throwError(err);
      });

      this.subscriptions.push(deleteSub);

    }).catch((err: HttpErrorResponse) => {

      this.loadingRequest = false;

      this.throwError(err);
    });
  }

  public rejectBlocklistRequest(item: BlocklistItem) {

    // Checker cannot be the same as maker
    if (item.approval.requestInitiatedBy === this.authService.userProfile.profile.email) {

      this.notificationService.show(
        'error',
        'Rejector of request cannot be the same as the requestor.',
        null,
        'Unauthorized Action',
        'top'
      );
      return;
    }

    this.loadingRequest = true;

    const approvalId = item.approval ? item.approval.key : null;

    // check to make sure approval exists
    if (approvalId && item.request.approvalId) {

      const deleteSub = this.portalApiService.deleteBlocklistRequest(item.approval.key).subscribe(() => {
        this.loadingRequest = false;
        this.getBlockedList();
      }, err => {
        this.throwError(err);
      });

      this.subscriptions.push(deleteSub);
    }
  }

  /**
   * Deletes blocklist request from firebase
   *
   * @param {BlocklistItem} item
   * @memberof BlocklistManagementComponent
   */
  public deleteBlocklistRequest(item: BlocklistItem) {

    if (item.approval.requestInitiatedBy === this.authService.userProfile.profile.email) {

      this.notificationService.show(
        'error',
        'Approver of request cannot be the same as the requestor.',
        null,
        'Unauthorized Action',
        'top'
      );
      return;
    }

    this.loadingRequest = true;

    this.blocklistService.removeFromBlocklist(item.request, 'approve')
      .then(() => {
        const deleteSub = this.portalApiService.deleteBlocklistRequest(item.approval.key).subscribe(() => {
          this.loadingRequest = false;
          this.getBlockedList();
        }, err => {
          this.throwError(err);
        });

        this.subscriptions.push(deleteSub);

      }).catch((err: HttpErrorResponse) => {
        this.loadingRequest = false;

        this.throwError(err);
      });
  }

  public openModal(action: string, type: BlocklistType, isoCode?: string, name?: string) {
    this.openBlocklistDialog(action, type, isoCode, name);
  }

  /**
   * Opens dialog for acting (add/delete) upon a blocklisted item
   *
   * @private
   * @param {string} action
   * @param {BlocklistType} type
   * @param {string} isoCode
   * @param {string} [name]
   * @memberof BlocklistManagementComponent
   */
  private openBlocklistDialog(action: string, type: BlocklistType, isoCode: string, name?: string) {

    const data: any = {
      action: action,
      type: type,
      isoCode: isoCode,
      name: name
    };

    const dialogRef = this.dialog.open(BlocklistDialogComponent, {
      disableClose: false,
      data: data
    });

    dialogRef.afterClosed().subscribe(result => this.getBlockedList());

  }

  private throwError(error: HttpErrorResponse) {

    const wwError: WorldWireError = error.error ? error.error : null;

    switch (error.status) {
      case 401:
      case 0: {
        this.notificationService.show(
          'error',
          'You are not authorized to perform this action.',
          null,
          'Unauthorized',
          'top'
        );
        break;
      }
      case 400: {
        const message = wwError && wwError.details ? wwError.details : 'Invalid form data was submitted and could not be processed. Please try again.';

        this.notificationService.show(
          'error',
          message,
          null,
          'Bad Request',
          'top'
        );
        break;
      }
      case 404: {
        const message = wwError && wwError.details ? wwError.details : 'Network could not be reached to make this request. Please contact administrator.';

        this.notificationService.show(
          'error',
          message,
          null,
          'Network Error',
          'top'
        );
        break;
      }
      default: {
        // for 500s and other errors: Unexpected ERROR notification
        this.notificationService.show(
          'error',
          'Unexpected error when making this request. Please contact administrator.',
          null,
          'Network Error',
          'top'
        );
        break;
      }
    }
  }
}
