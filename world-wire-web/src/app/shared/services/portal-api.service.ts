import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';

import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { IExchange, ITransaction } from '../models/transaction.interface';
import { WhitelistRequest } from '../../portal/shared/models/whitelist-request.interface';
import { Approval } from '../models/approval.interface';
import { IUserProfile } from '../models/user.interface';
import { IInstitution, IParticipantUsers } from '../models/participant.interface';
import { IJWTPublic } from '../models/token.interface';
import { TrustRequest } from '../../portal/shared/models/trust-request.interface';
import { AccountRequest } from '../models/account.interface';
import { SessionService } from './session.service';
import { AssetRequest } from '../models/asset.interface';
import { KillSwitchRequestDetail } from '../../portal/shared/models/killswitch-request.interface';
import { INodeAutomation } from '../models/node.interface';
import { BlocklistRequest } from '../models/blocklist.interface';

@Injectable()
export class PortalApiServiceHelper {
  getGenericHeader(token: string): HttpHeaders {
    return new HttpHeaders().set(
      'Authorization', `Bearer ${token}`).set(
      'Content-Type', 'application/json',
    );
  }
}

@Injectable()
export class PortalApiService {
  portalApiBaseUrl: string;

  constructor(private http: HttpClient, private sessionService: SessionService, private portalApiServiceHelper: PortalApiServiceHelper) {
    this.portalApiBaseUrl = environment.portalApiUrl;
  }

  getUserProfile(userId: string): Observable<UserPortalApiEntity> {
    const apiUrl  = `${this.portalApiBaseUrl}/users/${userId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<UserPortalApiEntity>(apiUrl, { headers });
  }

  getUserProfileP(userId: string): Promise<UserPortalApiEntity> {
    return this.getUserProfile(userId).toPromise();
  }

  getAllUserProfiles(institutionId: string): Observable<UserPortalApiEntity[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${institutionId}/users`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<UserPortalApiEntity[]>(apiUrl, { headers });
  }

  getInstitution(institutionId: string): Observable<IInstitution> {
    const apiUrl  = `${this.portalApiBaseUrl}/institutions/${institutionId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<IInstitution>(apiUrl, { headers });
  }

  getInstitutionP(institutionId: string): Promise<IInstitution> {
    return this.getInstitution(institutionId).toPromise();
  }

  createInstitution(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/institutions`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, requestBody, { headers });
  }

  createInstitutionP(requestBody: { [key: string]: any }): Promise<object> {
    return this.createInstitution(requestBody).toPromise();
  }

  updateInstitution(institutionId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/institutions/${institutionId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  updateInstitutionP(institutionId: string, requestBody: { [key: string]: any }): Promise<object> {
    return this.updateInstitution(institutionId, requestBody).toPromise();
  }

  deleteInstitution(institutionId: string): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/institutions/${institutionId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.delete(apiUrl, { headers });
  }

  deleteInstitutionP(institutionId: string): Promise<object> {
    return this.deleteInstitution(institutionId).toPromise();
  }

  getAllInstitution(): Observable<IInstitution[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/institutions`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<IInstitution[]>(apiUrl, { headers });
  }

  getTransactions(participantId: string, transactionType: string): Observable<ITransaction[] | IExchange[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/${transactionType}/transactions`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<ITransaction[] | IExchange[]>(apiUrl, { headers });
  }

  deleteParticipantPermission(institutionId: string, userId: string): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/${institutionId}/users/${userId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.delete(apiUrl, { headers });
  }

  deleteParticipantPermissionP(institutionId: string, userId: string): Promise<object> {
    return this.deleteParticipantPermission(institutionId, userId).toPromise();
  }

  getSuperPermissions(): Observable<IUserProfile[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/super/users`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<IUserProfile[]>(apiUrl, { headers });
  }

  updateSuperPermission(requestBody: { userId: string, role: string }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/super/users`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  deleteSuperPermission(userId: string): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/super/users/${userId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.delete(apiUrl, { headers });
  }

  deleteSuperPermissionP(userId: string): Promise<object> {
    return this.deleteSuperPermission(userId).toPromise();
  }

  getWhiteListRequests(participantId: string): Observable<WhitelistRequest[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/whitelist_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<WhitelistRequest[]>(apiUrl, { headers });
  }

  createWhiteListRequest(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/whitelist_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, requestBody, { headers });
  }

  updateWhiteListRequest(whitelistedId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/whitelist_requests/${whitelistedId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  deleteWhiteListRequest(approvalId): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/whitelist_requests/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.delete(apiUrl, { headers });
  }

  deleteWhiteListRequestP(approvalId): Promise<object> {
    return this.deleteWhiteListRequest(approvalId).toPromise();
  }

  getParticipantApproval(approvalId: string): Observable<Approval> {
    const apiUrl  = `${this.portalApiBaseUrl}/participant_approvals/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<Approval>(apiUrl, { headers });
  }

  getParticipantApprovalP(approvalId: string): Promise<Approval> {
    return this.getParticipantApproval(approvalId).toPromise();
  }

  updateParticipantApproval(approvalId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/participant_approvals/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  updateParticipantApprovalP(approvalId: string, requestBody: { [key: string]: any }): Promise<object> {
    return this.updateParticipantApproval(approvalId, requestBody).toPromise();
  }

  getSuperApproval(approvalId: string): Observable<Approval> {
    const apiUrl  = `${this.portalApiBaseUrl}/super_approvals/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<Approval>(apiUrl, { headers });
  }

  getSuperApprovalP(approvalId: string): Promise<Approval> {
    return this.getSuperApproval(approvalId).toPromise();
  }

  updateSuperApproval(approvalId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/super_approvals/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  getJWTInfoOfInstitution(institutionId: string): Observable<JwtInfoPortalApiEntity[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${institutionId}/jwt_info`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<JwtInfoPortalApiEntity[]>(apiUrl, { headers });
  }

  createTrustRequest(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/trust_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, requestBody, { headers });
  }

  updateTrustRequest(trustRequestId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/trust_requests/${trustRequestId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.patch(apiUrl, requestBody, { headers });
  }

  updateTrustRequestP(trustRequestId: string, requestBody): Promise<object> {
    return this.updateTrustRequest(trustRequestId, requestBody).toPromise();
  }

  getAllTrustRequests(participantId: string, requestField: string): Observable<TrustRequest[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/trust_requests/${requestField}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<TrustRequest[]>(apiUrl, { headers });
  }

  getAllAccountRequests(participantId: string): Observable<AccountRequest[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/account_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<AccountRequest[]>(apiUrl, { headers });
  }

  createAccountRequest(participantId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/account_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, { participantId, ...requestBody }, { headers });
  }

  updateAccountRequest(participantId: string, approvalId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/account_requests/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, { participantId, ...requestBody }, { headers });
  }

  getAllAssetRequests(participantId: string): Observable<AssetRequest[]> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/asset_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<AssetRequest[]>(apiUrl, { headers });
  }

  createAssetRequest(participantId: string, requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/asset_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, { participantId, ...requestBody }, { headers });
  }

  updateAssetRequest(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/asset_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers });
  }

  getKillSwitchRequest(participantId: string, accountAddress: string): Observable<KillSwitchRequestDetail | undefined> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/killswitch_requests/${accountAddress}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.get<KillSwitchRequestDetail | undefined>(apiUrl, { headers });
  }

  createKillSwitchRequest(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/killswitch_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, requestBody, { headers });
  }

  createKillSwitchRequestP(requestBody: { [key: string]: any }): Promise<object> {
    return this.createKillSwitchRequest(requestBody).toPromise();
  }

  updateKillSwitchRequest(participantId: string, accountAdress: string, requestBody: { [key: string]: any }): Promise<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/${participantId}/killswitch_requests/${accountAdress}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.put(apiUrl, requestBody, { headers }).toPromise();
  }

  getAllBlocklistRequests(type: string): Observable<BlocklistRequest[]> {
    const apiUrl      = `${this.portalApiBaseUrl}/blocklist_requests`;
    const headers     = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);
    const queryParams = new HttpParams().set('type', type);

    return this.http.get<BlocklistRequest[]>(apiUrl, { ...{ params: queryParams }, ...{ headers: headers } });
  }

  getAllBlocklistRequestsP(type: string): Promise<BlocklistRequest[]> {
    return this.getAllBlocklistRequests(type).toPromise();
  }

  createBlocklistRequest(requestBody: { [key: string]: any }): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/blocklist_requests`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.post(apiUrl, requestBody, { headers });
  }

  deleteBlocklistRequest(approvalId: string): Observable<object> {
    const apiUrl  = `${this.portalApiBaseUrl}/blocklist_requests/${approvalId}`;
    const headers = this.portalApiServiceHelper.getGenericHeader(this.sessionService.accessToken);

    return this.http.delete(apiUrl, { headers });
  }
}

export interface ParticipantPermissionPortalApiEntity {
  roles: {
    admin?: boolean,
    manager?: boolean,
    viewer?: boolean,
  };
  institution_id: string;
  user_id: string;
}

export interface UserPortalApiEntity {
  _id: string;
  profile: {
    email: string,
  };
  super_permission?: {
    roles: {
      admin?: boolean,
      manager?: boolean,
      viewer?: boolean,
    },
  };
  key?: string;
  registered?: boolean;
  participant_permissions?: ParticipantPermissionPortalApiEntity[];
}

export interface JwtInfoPortalApiEntity {
  institution: string;
  acc: string[];
  active: boolean;
  approvedAt: number;
  approvedBy: string;
  aud: string;
  createdAt: number;
  createdBy: string;
  description: string;
  enp: string[];
  env: string;
  ips: string[];
  jti: string;
  stage: 'review' | 'approved' | 'ready' | 'initialized' | 'revoked';
  sub: string;
  revokedAt: number;
  revokedBy: string;
  refreshedAt: number;
  ver: string;
}

@Injectable()
export class PortalApiServiceModelTransformer {
  transformUserPortalApiEntityToIParticipantUsers(userPortalApiEntities: UserPortalApiEntity[]): IParticipantUsers {
    const iParticipantUsers: IParticipantUsers = { users: {} };
    userPortalApiEntities.forEach((user) => {
      if (Array.isArray(user.participant_permissions)) {
        user.participant_permissions.forEach((permission) => {
          iParticipantUsers.users[permission.user_id] = {
            roles: permission.roles,
            profile: {
              email: user.profile.email,
            },
          };
        });
      }
    });

    return iParticipantUsers;
  }

  transformUserPortalApiEntityToIUserProfile(userPortalApiEntities: UserPortalApiEntity[]): IUserProfile[] {
    const iUserProfiles: IUserProfile[] = [];
    userPortalApiEntities.forEach((user) => {
      const iUserProfile: IUserProfile = {
        profile: { email: user.profile.email },
      };

      if (user.key) {
        iUserProfile['key'] = user.key;
      }

      if (user.registered) {
        iUserProfile['registered'] = user.registered;
      }

      if (user.super_permission) {
        iUserProfile['super_permission'] = { ...{ roles: user.super_permission.roles }, ...{ email: user.profile.email } };
      }

      if (Array.isArray(user.participant_permissions) && user.participant_permissions.length > 0) {
        iUserProfile.participant_permissions = {};

        user.participant_permissions.forEach((permission) => {
          iUserProfile.participant_permissions[permission.institution_id] = {
            email: user.profile.email,
            name: permission.institution_id,
            roles: permission.roles,
            slug: permission.institution_id,
          };
        });
      } else {
        iUserProfile.participant_permissions = {
          'institutionId': {
            roles: { admin: true },
            email: 'email',
            name: 'name',
            slug: 'slug',
          },
        };
      }
      iUserProfiles.push(iUserProfile);
    });

    return iUserProfiles;
  }

  transformJwtInfoPortalApiEntityToIJWTPublic(jwtInfoPortalApiEntities: JwtInfoPortalApiEntity[]): { [key: string]: IJWTPublic } {
    const iJWTPublicTokensObj: { [key: string]: IJWTPublic } = {};
    jwtInfoPortalApiEntities.forEach((jwtPortalEntity) => {
      iJWTPublicTokensObj[jwtPortalEntity.jti] = {
        acc: jwtPortalEntity.acc,
        active: jwtPortalEntity.active,
        approvedAt: jwtPortalEntity.approvedAt,
        approvedBy: jwtPortalEntity.approvedBy,
        aud: jwtPortalEntity.aud,
        createdAt: jwtPortalEntity.createdAt,
        createdBy: jwtPortalEntity.createdBy,
        enp: jwtPortalEntity.enp,
        env: jwtPortalEntity.env,
        ips: jwtPortalEntity.ips,
        jti: jwtPortalEntity.jti,
        refreshedAt: jwtPortalEntity.refreshedAt,
        stage: jwtPortalEntity.stage,
        ver: jwtPortalEntity.ver,
        description: jwtPortalEntity.description,
      };
    });

    return iJWTPublicTokensObj;
  }

  updateOrAddInstitutionNode(institution: IInstitution, nodeId: string, updateBody: INodeAutomation | { [key: string]: any }): IInstitution {
    if (!institution.nodes) {
      institution.nodes = [];
    }
    const foundNode = institution.nodes.some((institutionNode, index) => {
      if (institutionNode.participantId === nodeId) {
        institution.nodes[index] = {
          ...institution.nodes[index],
          ...updateBody,
        };
        return true;
      }
    });

    if (!foundNode) {
      institution.nodes.push(updateBody as INodeAutomation);
    }
    return institution;
  }

  getInstitutionNode(institution: IInstitution, nodeId: string): INodeAutomation | undefined {
    return institution.nodes.find((node) => {
      return node.participantId === nodeId;
    });
  }
}
