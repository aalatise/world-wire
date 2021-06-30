import * as _ from 'lodash';
import { ENVIRONMENT } from '../constants/general.constants';
import { Injectable, isDevMode } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';

import { Blocklist, BlocklistRequest, BlocklistType } from '../models/blocklist.interface';
import { AuthService } from './auth.service';
import { ApprovalPermission } from '../models/approval.interface';
import { environment } from '../../../environments/environment';
import { PortalApiService } from './portal-api.service';


@Injectable()
export class BlocklistService {

  private adminBlocklistUrl = `${environment.adminApiUrl}/blocklist`;

  constructor(
    public http: HttpClient,
    private authService: AuthService,
    private portalApiService: PortalApiService,
  ) { }

  /**
   * Get a list of values from the blocklist.
   *
   * @param {('country' | 'currency' | 'institution')} type
   * @returns {Promise<string[]>}
   * @memberof BlocklistService
   */
  public async getBlocklist(blocklistType: BlocklistType) {

    try {
      let headers: HttpHeaders = this.authService.createSuperUserHeaders();

      const options = {
        headers
      };

      const blocklist: Blocklist = await this.http.get(
        `${this.adminBlocklistUrl}?type=${blocklistType}`,
        options
      ).toPromise() as Blocklist;

      return blocklist[0].value as string[] | [];

    } catch (err) {
      if (isDevMode()) {
        console.log(err);
      }
      return [];
    }
  }

  /**
   * Add to blocklist/post request
   *
   * @memberof BlocklistService
   */
  public async addToBlocklist(request: BlocklistRequest, permission: ApprovalPermission): Promise<any> {


      const approvalId = request.approvalId ? request.approvalId: null;

      let headers: HttpHeaders = this.authService.createSuperUserHeaders();

      if (permission === 'request') {
        headers = this.authService.addMakerCheckerHeaders(headers, permission);
      } else {
        headers = this.authService.addMakerCheckerHeaders(headers, permission, approvalId);
      }

      const options = {
        headers
      };

      const body: Blocklist = {
        type: request.type,
        value: [request.value]
      };

      return this.http.post(
        this.adminBlocklistUrl,
        body,
        options
      ).toPromise();
  }

  /**
   * Sends API request to remove currency, country,
   * institution from Blocklist.
   *
   * @param {BlocklistRequest} request
   * @param {ApprovalPermission} permission
   * @returns {Promise<any>}
   * @memberof BlocklistService
   */
  public async removeFromBlocklist(request: BlocklistRequest, permission: ApprovalPermission): Promise<any> {
      const approvalId = request.approvalId ? request.approvalId : null;

      let headers: HttpHeaders = await this.authService.createSuperUserHeaders();

      if (permission === 'request') {
        headers = this.authService.addMakerCheckerHeaders(headers, permission);
      } else {
        headers = this.authService.addMakerCheckerHeaders(headers, permission, approvalId);
      }

      const body: Blocklist = {
        type: request.type,
        value: [request.value]
      };

      // Per Angular 7+ spec, body for DELETE can be included in options
      const options = {
        headers,
        body: body
      };

      return this.http.delete(
        this.adminBlocklistUrl,
        options
      ).toPromise();

  }

  /**
   * Gets all blocklist requests for approval data
   *
   * @returns {Promise<BlocklistRequest>}
   * @memberof BlocklistService
   */
  public getBlocklistRequests(type: BlocklistType): Promise<BlocklistRequest[]> {
    return this.portalApiService.getAllBlocklistRequestsP(type);
  }
}
